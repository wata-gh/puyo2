#!/usr/bin/env python3
"""Bounded Agents API maintenance controller. Python 3.11+, no SDK dependency."""
import argparse
import copy
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import signal
import subprocess
import tarfile
import tempfile
import time
import tomllib
import urllib.error
import urllib.parse
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[2]
MANIFEST = 'crates/puyo2/Cargo.toml'
TOOLCHAIN = 'rust-toolchain.toml'


class Failure(RuntimeError):
    pass


def scrub(text):
    for name in ('OPENAI_API_KEY', 'OPENAI_EXECUTOR_API_KEY', 'GH_TOKEN', 'GITHUB_TOKEN'):
        value = os.environ.get(name)
        if value:
            text = text.replace(value, '[REDACTED]')
    return re.sub(r'(?:sk-[A-Za-z0-9_-]+|gh[opsu]_[A-Za-z0-9_]+|github_pat_[A-Za-z0-9_]+)', '[REDACTED]', text)


def command(args, *, cwd=None, data=None, timeout=120, env=None, combined=False):
    # No shell; capture output so secrets can be redacted before any persistence.
    try:
        p = subprocess.run(args, cwd=cwd, input=data, stdout=subprocess.PIPE,
                           stderr=subprocess.STDOUT if combined else subprocess.PIPE, timeout=timeout, env=env)
    except subprocess.TimeoutExpired:
        raise Failure('command timed out: ' + args[0]) from None
    if p.returncode:
        raise Failure(scrub('command failed: ' + ' '.join(args) + '\n' + (p.stderr or b'').decode(errors='replace')[-12000:] + p.stdout.decode(errors='replace')[-12000:]))
    return p.stdout


def http(url, *, method='GET', body=None, key=None, beta=False):
    headers = {'User-Agent': 'puyo2-maintenance', 'Content-Type': 'application/json'}
    if key:
        headers['Authorization'] = 'Bearer ' + key
    if beta:
        headers['OpenAI-Beta'] = 'agents=v1'
    req = urllib.request.Request(url, data=None if body is None else json.dumps(body).encode(), headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as response:
            raw = response.read(4 * 1024 * 1024)
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        # Never persist response bodies, headers, or credentials.
        raise Failure(f'HTTP {e.code}: {method} {urllib.parse.urlsplit(url).path}') from None
    except (OSError, ValueError):
        raise Failure(f'API transport/JSON failure: {method} {urllib.parse.urlsplit(url).path}') from None


def version(value):
    if not re.fullmatch(r'\d+\.\d+(?:\.\d+)?', value):
        raise Failure('expected stable numeric version')
    return tuple(map(int, (value + '.0').split('.')[:3]))


def archive_files(raw, limit):
    """Read a Docker/git tar as data, never extract untrusted paths or links."""
    if len(raw) > limit:
        raise Failure('archive size limit')
    result = {}
    total = 0
    with tarfile.open(fileobj=io.BytesIO(raw)) as tar:
        for item in tar:
            parts = PurePosixPath(item.name).parts
            if parts and parts[0] == 'workspace':
                parts = parts[1:]
            if not parts or item.isdir():
                continue
            if item.name.startswith('/') or '..' in parts or not item.isfile():
                raise Failure('unsafe archive member')
            name = '/'.join(parts)
            if name in result:
                raise Failure('duplicate archive member')
            total += item.size
            if total > limit:
                raise Failure('expanded archive size limit')
            # Git tracks only the executable bit, while git archive's tar.umask
            # may add group-write bits (0664/0775). Canonicalize both Git and
            # Docker snapshots before comparisons and repacking.
            mode = 0o755 if item.mode & 0o100 else 0o644
            result[name] = (tar.extractfile(item).read(), mode)
    return result


def pack(files):
    buffer = io.BytesIO()
    with tarfile.open(fileobj=buffer, mode='w') as tar:
        directories = sorted({str(parent) for name in files for parent in PurePosixPath(name).parents if str(parent) != '.'})
        for name in directories:
            item = tarfile.TarInfo(name)
            item.type, item.mode, item.uid, item.gid = tarfile.DIRTYPE, 0o755, 10001, 10001
            tar.addfile(item)
        for name, (data, mode) in sorted(files.items()):
            item = tarfile.TarInfo(name)
            item.size, item.mode, item.uid, item.gid = len(data), mode, 10001, 10001
            tar.addfile(item, io.BytesIO(data))
    return buffer.getvalue()


def validate_diff(base, proposed, kind):
    """Conservative gate: protect tests, workflows, public metadata and MSRV."""
    if set(base) - set(proposed):
        raise Failure('file deletion is forbidden')
    changed = [p for p in proposed if proposed[p] != base.get(p)]
    for name in ('OPENAI_API_KEY', 'OPENAI_EXECUTOR_API_KEY', 'GH_TOKEN', 'GITHUB_TOKEN'):
        secret = os.environ.get(name, '').encode()
        if secret and any(secret in proposed[p][0] for p in changed):
            raise Failure('secret detected in candidate')
    source_changed = any(p.startswith('crates/puyo2/src/') for p in changed)
    if source_changed and not any(p.startswith('crates/puyo2/tests/') and p not in base for p in changed):
        raise Failure('source fixes require a new regression test')
    for path in changed:
        old, old_mode = base.get(path, (b'', 0o644))
        new, mode = proposed[path]
        if mode != old_mode or mode not in (0o644, 0o755):
            raise Failure('file mode change is forbidden')
        allowed = path in ('Cargo.lock', MANIFEST) or (kind == 'rust' and path == TOOLCHAIN)
        if path.startswith('crates/puyo2/src/') and path.endswith('.rs'):
            allowed = True
            # Existing inline test modules must be byte-identical. Any other test
            # syntax is conservatively blocked for human review, not auto-approved.
            marker = b'#[cfg(test)]'
            if marker in old:
                if old[old.index(marker):] != new[new.find(marker):]:
                    raise Failure('inline tests changed')
            elif b'#[test]' in old or b'#[test]' in new:
                raise Failure('test-bearing source needs human review')
            if re.search(rb'#\s*\[\s*(?:ignore|cfg|cfg_attr)\b', new.replace(old, b'')):
                # Compare attribute lines separately below to permit unchanged cfg.
                attrs = lambda b: re.findall(rb'^.*#\s*\[\s*(?:ignore|cfg|cfg_attr)\b.*$', b, re.M)
                if attrs(old) != attrs(new):
                    raise Failure('conditional compilation change')
        if path.startswith('crates/puyo2/tests/') and path.endswith('.rs') and path not in base:
            allowed = True  # New regression tests only.
        if not allowed:
            raise Failure('protected path changed: ' + path)
    old = tomllib.loads(base[MANIFEST][0].decode())
    new = tomllib.loads(proposed[MANIFEST][0].decode())
    old_deps, new_deps = old.pop('dependencies'), new.pop('dependencies')
    if old != new or set(old_deps) != set(new_deps):
        raise Failure('package metadata, MSRV, edition or dependency set changed')
    for name in old_deps:
        a, b = old_deps[name], new_deps[name]
        a = {'version': a} if isinstance(a, str) else copy.deepcopy(a)
        b = {'version': b} if isinstance(b, str) else copy.deepcopy(b)
        av, bv = a.pop('version'), b.pop('version')
        if a != b or version(bv) < version(av):
            raise Failure('dependency features/source/downgrade changed')
    a = tomllib.loads(base[TOOLCHAIN][0].decode())
    b = tomllib.loads(proposed[TOOLCHAIN][0].decode())
    av, bv = a['toolchain'].pop('channel'), b['toolchain'].pop('channel')
    if a != b or version(bv) < version(av) or (kind != 'rust' and av != bv):
        raise Failure('toolchain policy violation')
    lock = tomllib.loads(proposed['Cargo.lock'][0].decode())
    for package in lock['package']:
        if package.get('source') and package['source'] != 'registry+https://github.com/rust-lang/crates.io-index':
            raise Failure('non crates.io dependency source')
    return sorted(changed)


def detect(base, kind, fetch=http):
    result = copy.deepcopy(base)
    notes = []
    if kind == 'rust':
        # Official stable channel manifest; no inference from release dates.
        url = 'https://static.rust-lang.org/dist/channel-rust-stable.toml'
        with urllib.request.urlopen(url, timeout=30) as response:
            latest = tomllib.loads(response.read().decode())['pkg']['rust']['version'].split()[0]
        current = tomllib.loads(base[TOOLCHAIN][0].decode())['toolchain']['channel']
        if version(latest) > version(current):
            text = base[TOOLCHAIN][0].decode().replace(f'channel = "{current}"', f'channel = "{latest}"')
            result[TOOLCHAIN] = (text.encode(), base[TOOLCHAIN][1])
        notes.append(f'Rust runtime {current} -> {latest}; MSRV and edition unchanged')
    else:
        manifest = tomllib.loads(base[MANIFEST][0].decode())
        msrv = version(manifest['package']['rust-version'])
        text = base[MANIFEST][0].decode()
        for name, spec in manifest['dependencies'].items():
            current = spec if isinstance(spec, str) else spec['version']
            releases = fetch('https://crates.io/api/v1/crates/' + name)['versions']
            stable = [r for r in releases if not r['yanked'] and re.fullmatch(r'\d+\.\d+\.\d+', r['num'])]
            eligible = [r for r in stable if not r.get('rust_version') or version(r['rust_version']) <= msrv]
            if not eligible:
                raise Failure('no MSRV-compatible candidate for ' + name)
            latest = max(eligible, key=lambda r: version(r['num']))['num']
            if version(latest) > version(current):
                pattern = r'(?m)^(' + re.escape(name) + r'\s*=\s*(?:\{\s*version\s*=\s*)?")' + re.escape(current) + r'(")'
                text, count = re.subn(pattern, lambda m: m[1] + latest + m[2], text)
                if count != 1:
                    raise Failure('unsupported manifest formatting')
            notes.append(f'{name}: {current} -> {latest} (latest compatible with declared MSRV)')
        result[MANIFEST] = (text.encode(), base[MANIFEST][1])
    return result, notes


class Sandbox:
    def __init__(self, config, files):
        self.config = config
        self.name = 'puyo2-maint-' + uuid.uuid4().hex
        self.files = files

    def __enter__(self):
        try:
            command(['docker', 'run', '-d', '--name', self.name, '--label', 'puyo2-maintenance=true',
                     '--cap-drop=ALL', '--security-opt=no-new-privileges', '--pids-limit=256',
                     '--memory=6g', '--cpus=2', self.config['image']])
            command(['docker', 'cp', '-a', '-', self.name + ':/workspace'], data=pack(self.files))
            return self
        except BaseException:
            self.__exit__(None, None, None)
            raise

    def run(self, args):
        return command(['docker', 'exec', self.name, *args], timeout=self.config['command_seconds'], combined=True).decode(errors='replace')

    def snapshot(self):
        # Read a bounded stream from Docker daemon, not a shell/tar controlled by agent.
        with subprocess.Popen(['docker', 'cp', self.name + ':/workspace/.', '-'], stdout=subprocess.PIPE, stderr=subprocess.DEVNULL) as p:
            raw = p.stdout.read(self.config['max_archive_bytes'] + 1)
            if len(raw) > self.config['max_archive_bytes']:
                p.kill()
                raise Failure('workspace archive too large')
            if p.wait(timeout=30):
                raise Failure('workspace export failed')
        return archive_files(raw, self.config['max_archive_bytes'])

    def __exit__(self, *_):
        command(['docker', 'rm', '-f', '-v', self.name], timeout=30)


class Agents:
    def __init__(self, config, report):
        self.config, self.report = config, report
        self.session = None
        self.active = False

    def call(self, method, suffix='', body=None):
        return http('https://api.openai.com/v1/agents/sessions' + suffix, method=method,
                    body=body, key=os.environ['OPENAI_API_KEY'], beta=True)

    def __enter__(self):
        self.session = self.call('POST', body={
            'agent': {'model': self.config['model'], 'multi_agent': {'enabled': False},
                      'instructions': 'Maintain only puyo2 in /workspace. Follow AGENTS.md. Do not weaken or delete tests. Do not change MSRV, edition, CI, features, public APIs, or unrelated code. No git commits, network publication, releases or subagents. Build artifacts belong in /build. Apply only necessary compatibility fixes and add regression tests. Finish within this turn.'},
            'environment': {'type': 'self_hosted', 'workspace_directory': '/workspace'},
            'metadata': {'purpose': 'puyo2-maintenance', 'run': self.report['run']}})
        self.report['session_id'] = self.session['id']
        Path(self.report['session_file']).write_text(json.dumps({'id': self.session['id']}))
        return self

    def connect(self, sandbox):
        env = self.session['environment']
        remote = urllib.parse.urlsplit(env['remote_url'])
        if remote.scheme != 'https' or remote.hostname != 'api.openai.com':
            raise Failure('unexpected executor registration URL')
        # Only the restricted environment key crosses the container boundary.
        child_env = {**os.environ, 'CODEX_API_KEY': os.environ['OPENAI_EXECUTOR_API_KEY']}
        command(['docker', 'exec', '-d', '-e', 'CODEX_API_KEY', sandbox.name, 'codex', 'exec-server',
                 '--remote', env['remote_url'], '--environment-id', env['id']], env=child_env)

    def turns(self):
        result, cursor = [], ''
        while True:
            page = self.call('GET', '/' + self.session['id'] + '/turns?limit=100&order=asc' + cursor)
            result.extend(page['data'])
            if not page['has_more']:
                return result
            cursor = '&after=' + urllib.parse.quote(page['last_id'], safe='')
            if len(result) > 100:
                raise Failure('unexpected number of agent turns')

    def repair(self, message):
        before = {t['id'] for t in self.turns()}
        self.active = True  # A lost POST response may still have started work.
        self.call('POST', '/' + self.session['id'] + '/events', {'events': [{
            'type': 'agent.session.input.message',
            'input': [{'role': 'user', 'content': [{'type': 'input_text', 'text': message}]}]}]})
        deadline = time.monotonic() + self.config['turn_seconds']
        unknown_since = time.monotonic()
        while time.monotonic() < deadline:
            turns = self.turns()
            total = sum(t['usage']['total_tokens'] for t in turns if t.get('usage'))
            self.report['observed_tokens'] = total
            self.report['turns'] = [{'id': t['id'], 'status': t['status'], 'usage': t.get('usage'), 'error': t.get('error')} for t in turns]
            if total >= self.config['max_observed_tokens']:
                raise Failure('observed token ceiling reached')
            unknown = not turns or any(t.get('usage') is None for t in turns)
            if unknown and time.monotonic() - unknown_since > self.config['usage_grace_seconds']:
                raise Failure('usage unavailable; cannot enforce usage watchdog')
            if not unknown:
                unknown_since = time.monotonic()
            new = [t for t in turns if t['id'] not in before and not t.get('subagent_id')]
            if any(t.get('subagent_id') for t in turns):
                raise Failure('unexpected subagent')
            if any(t['status'] in ('failed', 'cancelled') for t in new):
                raise Failure('agent turn failed or cancelled')
            if new and all(t['status'] == 'completed' for t in new) and not unknown:
                self.active = False
                return
            state = self.call('GET', '/' + self.session['id'])
            self.report['last_session_state'] = {'status': state['status'], 'error': state.get('error'),
                                                  'actions': [a['type'] for a in state.get('required_actions', [])]}
            if state['status'] == 'failed' or any(a['type'] != 'environment_connection' for a in state.get('required_actions', [])):
                raise Failure('session failed or requires unsupported action')
            time.sleep(5)
        raise Failure('agent turn deadline exceeded')

    def __exit__(self, *_):
        if self.session:
            try:
                if self.active:
                    self.call('POST', '/' + self.session['id'] + '/events', {'events': [{'type': 'agent.session.input.cancel'}]})
            finally:
                self.call('DELETE', '/' + self.session['id'])
                self.report['session_deleted'] = True
                Path(self.report['session_file']).unlink(missing_ok=True)


def verify(config, files, report):
    with Sandbox(config, files) as box:
        try:
            output = box.run(['bash', 'ops/maintenance/verify.sh'])
            report.setdefault('validation', []).append({'result': 'passed', 'checks': [line for line in output.splitlines() if line.startswith(('CHECK:', 'test result:', 'checked='))], 'output': scrub(output)[-24000:]})
            return None
        except Failure as e:
            report.setdefault('validation', []).append({'result': 'failed', 'output': scrub(str(e))})
            return scrub(str(e))


def make_patch(base, result):
    with tempfile.TemporaryDirectory(prefix='puyo2-patch-') as tmp:
        root = Path(tmp)
        command(['git', 'init', '-q', tmp])
        for files in (base, result):
            for name, (data, mode) in files.items():
                p = root / name
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_bytes(data)
                p.chmod(mode)
            command(['git', 'add', '.'], cwd=root)
            if files is base:
                command(['git', '-c', 'user.name=Maintenance', '-c', 'user.email=maintenance@localhost', 'commit', '--no-gpg-sign', '-qm', 'Base'], cwd=root)
        return command(['git', 'diff', '--cached', '--binary', 'HEAD'], cwd=root)


def github_get(path):
    return http('https://api.github.com/repos/wata-gh/puyo2/' + path,
                key=os.environ.get('GH_TOKEN') or os.environ.get('GITHUB_TOKEN'))


def existing_pr(kind):
    branch = 'codex/maintenance-' + kind
    return github_get('pulls?state=open&base=main&head=' + urllib.parse.quote('wata-gh:' + branch))


def execute(args, config, report):
    base_sha = command(['git', 'rev-parse', 'HEAD'], cwd=ROOT).decode().strip()
    report.update(base_sha=base_sha, kind=args.kind, mode=args.mode)
    if args.mode != 'offline' and os.environ.get('GITHUB_ACTIONS') != 'true':
        raise Failure('live modes run only on disposable GitHub Actions runners; use --mode offline locally')
    if args.mode != 'offline' and existing_pr(args.kind):
        report['status'] = 'skipped_existing_pr'
        return
    if args.mode != 'offline':
        branch = 'refs/heads/codex/maintenance-' + args.kind
        if any(r['ref'] == branch for r in github_get('git/matching-refs/heads/codex/maintenance-' + args.kind)):
            raise Failure('orphan update branch exists; inspect before retry')
    base = archive_files(command(['git', 'archive', 'HEAD'], cwd=ROOT), config['max_archive_bytes'])
    if args.mode == 'offline':
        candidate, notes = copy.deepcopy(base), ['Offline: validate current committed tree; no APIs or publication']
    else:
        candidate, notes = detect(base, args.kind)
    report['candidates'] = notes
    with Sandbox(config, candidate) as box:
        preparation_failure = None
        if args.mode != 'offline' and args.kind == 'dependencies':
            try:
                box.run(['cargo', 'update'])
            except Failure as e:
                preparation_failure = str(e)
        candidate = box.snapshot()
        changed = validate_diff(base, candidate, args.kind)
        if not changed and not preparation_failure and args.mode != 'offline':
            report['status'] = 'no_updates'
            return
    failure = preparation_failure or verify(config, candidate, report)
    if failure and args.mode != 'offline':
        for key in ('OPENAI_API_KEY', 'OPENAI_EXECUTOR_API_KEY'):
            if not os.environ.get(key):
                raise Failure(key + ' is required for compatibility repair')
        with Agents(config, report) as agent, Sandbox(config, candidate) as box:
            agent.connect(box)
            for attempt in range(config['max_repairs']):
                report['repairs'] = attempt + 1
                agent.repair('Fix this update with minimal compatible changes. Existing tests are immutable. Add a regression test for any behavior fix.\nCandidates:\n' + '\n'.join(notes) + '\nTrusted validation failed:\n' + failure)
                candidate = box.snapshot()
                validate_diff(base, candidate, args.kind)
                failure = verify(config, candidate, report)
                if not failure:
                    break
    if failure:
        raise Failure('validation failed; repair budget exhausted or offline mode')
    patch = make_patch(base, candidate)
    if len(patch) > config['max_patch_bytes']:
        raise Failure('patch size limit')
    report.update(status='verified', changed=validate_diff(base, candidate, args.kind), patch_sha256=hashlib.sha256(patch).hexdigest())
    (args.output / 'update.patch').write_bytes(patch)
    report['downstream'] = 'puyo-rsrch-engine not executed or modified; public API compatibility requires review'


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--kind', choices=['dependencies', 'rust'], default='dependencies')
    parser.add_argument('--mode', choices=['offline', 'dry-run', 'publish'], default='dry-run')
    parser.add_argument('--output', type=Path, default=Path('/tmp/puyo2-maintenance-output'))
    args = parser.parse_args()
    config = json.loads((ROOT / 'ops/maintenance/config.json').read_text())
    args.output.mkdir(parents=True, exist_ok=True)
    # Never let a stale verified patch survive a failed rerun.
    for name in ('update.patch', 'report.json'):
        (args.output / name).unlink(missing_ok=True)
    report = {'run': os.environ.get('GITHUB_RUN_ID', uuid.uuid4().hex), 'status': 'failed', 'session_file': str(args.output / 'session.json')}
    def stop(*_):
        raise Failure('execution interrupted or overall deadline exceeded')
    signal.signal(signal.SIGTERM, stop)
    signal.signal(signal.SIGINT, stop)
    signal.signal(signal.SIGALRM, stop)
    signal.alarm(config['max_seconds'])
    try:
        execute(args, config, report)
    except Exception as e:
        report.update(status='failed', reason=scrub(str(e)))
        (args.output / 'update.patch').unlink(missing_ok=True)
    finally:
        signal.alarm(0)
        (args.output / 'report.json').write_text(scrub(json.dumps(report, indent=2)) + '\n')
    print('Maintenance result: ' + report['status'])
    return 1 if report['status'] == 'failed' else 0


if __name__ == '__main__':
    raise SystemExit(main())
