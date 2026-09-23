import copy
import io
import json
import os
from pathlib import Path
import sys
import tarfile
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import maintenance as m


def fixture():
    return {
        m.MANIFEST: (b'[package]\nname="puyo2"\nversion="0.2.0"\nrust-version="1.94"\nedition="2024"\n[dependencies]\nserde = { version = "1.0", features = ["derive"] }\n', 0o644),
        m.TOOLCHAIN: (b'[toolchain]\nchannel = "1.94.0"\ncomponents = ["rustfmt"]\n', 0o644),
        'Cargo.lock': (b'version=4\n[[package]]\nname="serde"\nversion="1.0.0"\nsource="registry+https://github.com/rust-lang/crates.io-index"\n', 0o644),
        'crates/puyo2/tests/existing.rs': (b'#[test]\nfn contract() { assert!(true); }', 0o644),
        'crates/puyo2/src/lib.rs': (b'pub fn value() -> u32 { 1 }\n#[cfg(test)]\nmod tests { #[test] fn value() { assert_eq!(super::value(), 1); } }', 0o644),
        '.github/workflows/rust.yml': (b'original CI', 0o644),
    }


class PolicyTests(unittest.TestCase):
    def setUp(self):
        self.base = fixture()
        self.new = copy.deepcopy(self.base)

    def change(self, path, old, new):
        b, mode = self.new[path]
        self.new[path] = (b.replace(old, new), mode)

    def test_dependency_version_allowed(self):
        self.change(m.MANIFEST, b'"1.0"', b'"1.0.228"')
        self.assertEqual(m.validate_diff(self.base, self.new, 'dependencies'), [m.MANIFEST])

    def test_metadata_features_msrv_edition_forbidden(self):
        for old, new in [(b'"1.94"', b'"1.95"'), (b'"2024"', b'"2021"'), (b'"derive"', b'"std"'), (b'"0.2.0"', b'"0.3.0"')]:
            with self.subTest(new=new):
                self.new = copy.deepcopy(self.base)
                self.change(m.MANIFEST, old, new)
                with self.assertRaises(m.Failure):
                    m.validate_diff(self.base, self.new, 'dependencies')

    def test_test_deletion_modification_workflow_forbidden(self):
        for path in ['crates/puyo2/tests/existing.rs', '.github/workflows/rust.yml']:
            for delete in [True, False]:
                self.new = copy.deepcopy(self.base)
                if delete:
                    del self.new[path]
                else:
                    self.new[path] = (b'weakened', 0o644)
                with self.assertRaises(m.Failure):
                    m.validate_diff(self.base, self.new, 'dependencies')

    def test_source_requires_regression_and_preserves_inline_tests(self):
        self.change('crates/puyo2/src/lib.rs', b'{ 1 }', b'{ 2 }')
        with self.assertRaises(m.Failure):
            m.validate_diff(self.base, self.new, 'dependencies')
        self.new['crates/puyo2/tests/regression.rs'] = (b'#[test] fn regression() {}', 0o644)
        m.validate_diff(self.base, self.new, 'dependencies')
        self.change('crates/puyo2/src/lib.rs', b'assert_eq!', b'debug_assert_eq!')
        with self.assertRaises(m.Failure):
            m.validate_diff(self.base, self.new, 'dependencies')

    def test_runtime_update_does_not_raise_msrv(self):
        self.change(m.TOOLCHAIN, b'1.94.0', b'1.95.0')
        m.validate_diff(self.base, self.new, 'rust')
        with self.assertRaises(m.Failure):
            m.validate_diff(self.base, self.new, 'dependencies')

    def test_secret_rejected(self):
        with patch.dict(os.environ, {'OPENAI_EXECUTOR_API_KEY': 'test-secret'}):
            self.new['crates/puyo2/tests/leak.rs'] = (b'test-secret', 0o644)
            with self.assertRaisesRegex(m.Failure, 'secret'):
                m.validate_diff(self.base, self.new, 'dependencies')

    def test_archive_rejects_escape_symlink_duplicate_and_limit(self):
        for name, link in [('../escape', False), ('/absolute', False), ('link', True)]:
            buffer = io.BytesIO()
            with tarfile.open(fileobj=buffer, mode='w') as tar:
                member = tarfile.TarInfo(name)
                if link:
                    member.type, member.linkname = tarfile.SYMTYPE, '/etc/passwd'
                tar.addfile(member)
            with self.assertRaises(m.Failure):
                m.archive_files(buffer.getvalue(), 20000)
        with self.assertRaises(m.Failure):
            m.archive_files(m.pack(self.base), 10)
        self.assertEqual(m.archive_files(m.pack(self.base), 100000), self.base)

    def test_new_dependency_sources_forbidden(self):
        self.change('Cargo.lock', b'registry+https://github.com/rust-lang/crates.io-index', b'git+https://example.invalid/evil')
        with self.assertRaises(m.Failure):
            m.validate_diff(self.base, self.new, 'dependencies')

    def test_detection_skips_yanked_prerelease_and_high_msrv(self):
        releases = [{'num': v, 'yanked': y, 'rust_version': r} for v, y, r in [
            ('1.0.9', False, '1.80'), ('1.0.10', True, '1.80'), ('2.0.0', False, '1.95'), ('3.0.0-beta.1', False, None)]]
        result, _ = m.detect(self.base, 'dependencies', lambda _: {'versions': releases})
        self.assertIn(b'1.0.9', result[m.MANIFEST][0])
        m.validate_diff(self.base, result, 'dependencies')

    def test_patch_can_be_applied_and_contains_new_tests(self):
        self.new['crates/puyo2/tests/new.rs'] = (b'new test', 0o644)
        patch_data = m.make_patch(self.base, self.new)
        self.assertIn(b'new file mode 100644', patch_data)
        self.assertIn(b'+new test', patch_data)


class LifecycleTests(unittest.TestCase):
    def setUp(self):
        self.config = {'turn_seconds': 10, 'usage_grace_seconds': 2, 'max_observed_tokens': 100}
        self.agent = m.Agents(self.config, {})
        self.agent.session = {'id': 'sess_test'}
        self.calls = []
        self.agent.call = lambda *args, **kwargs: self.calls.append(args) or {'status': 'idle'}

    def turn(self, status='completed', usage=10):
        return {'id': 'turn_1', 'subagent_id': None, 'status': status,
                'usage': None if usage is None else {'total_tokens': usage}}

    def test_completed_turn_accepted_with_usage(self):
        with patch.object(self.agent, 'turns', side_effect=[[], [self.turn()]]):
            self.agent.repair('fix')
        self.assertEqual(self.agent.report['observed_tokens'], 10)
        self.assertEqual(self.calls[0][1], '/sess_test/events')

    def test_failed_cancelled_over_budget_are_not_success(self):
        for turn in [self.turn('failed'), self.turn('cancelled'), self.turn(usage=101)]:
            with patch.object(self.agent, 'turns', side_effect=[[], [turn]]):
                with self.assertRaises(m.Failure):
                    self.agent.repair('fix')

    def test_idle_or_missing_usage_never_counts_as_success(self):
        for turns in [[], [self.turn(usage=None)]]:
            with patch.object(self.agent, 'turns', return_value=turns), patch.object(m.time, 'sleep'), patch.object(m.time, 'monotonic', side_effect=[0, 0, 1, 4]):
                with self.assertRaisesRegex(m.Failure, 'usage unavailable'):
                    self.agent.repair('fix')

    def test_timeout(self):
        with patch.object(self.agent, 'turns', return_value=[]), patch.object(m.time, 'monotonic', side_effect=[0, 0, 11]):
            with self.assertRaisesRegex(m.Failure, 'deadline'):
                self.agent.repair('fix')

    def test_cleanup_delete_even_if_cancel_fails(self):
        calls = []
        def fail_cancel(method, *args):
            calls.append(method)
            if method == 'POST':
                raise m.Failure('cancel failed')
        self.agent.call = fail_cancel
        self.agent.active = True
        with tempfile.TemporaryDirectory() as tmp:
            self.agent.report['session_file'] = str(Path(tmp) / 'session.json')
            with self.assertRaises(m.Failure):
                self.agent.__exit__(None, None, None)
        self.assertEqual(calls, ['POST', 'DELETE'])
        self.assertTrue(self.agent.report['session_deleted'])

    def test_turn_pagination(self):
        self.agent.call = lambda *args: {'data': [self.turn()], 'has_more': False} if 'after=' in args[1] else {'data': [self.turn()], 'has_more': True, 'last_id': 'one'}
        self.assertEqual(len(self.agent.turns()), 2)

    def test_application_key_never_sent_to_sandbox(self):
        self.agent.session['environment'] = {'id': 'env_1', 'remote_url': 'https://api.openai.com/v1/agents/api'}
        with patch.dict(os.environ, {'OPENAI_API_KEY': 'app-key', 'OPENAI_EXECUTOR_API_KEY': 'executor-key'}), patch.object(m, 'command') as cmd:
            self.agent.connect(type('Box', (), {'name': 'box'})())
        args = cmd.call_args.args[0]
        self.assertEqual(args[args.index('-e') + 1], 'CODEX_API_KEY')
        self.assertNotIn('app-key', args)
        self.assertNotIn('executor-key', args)

    def test_duplicate_skips_before_detection_or_agent(self):
        args = type('Args', (), {'mode': 'dry-run', 'kind': 'rust'})()
        report = {}
        with patch.dict(os.environ, {'GITHUB_ACTIONS': 'true'}), patch.object(m, 'command', return_value=b'abc'), patch.object(m, 'existing_pr', return_value=[{'number': 1}]), patch.object(m, 'detect') as detect:
            m.execute(args, {}, report)
        self.assertEqual(report['status'], 'skipped_existing_pr')
        detect.assert_not_called()

    def test_redacts_keys(self):
        with patch.dict(os.environ, {'OPENAI_API_KEY': 'unique-key'}):
            self.assertNotIn('unique-key', m.scrub('unique-key sk-123test ghp_123test'))
            self.assertNotIn('sk-', m.scrub('unique-key sk-123test ghp_123test'))

class ApiContractTests(unittest.TestCase):
    def test_beta_header_and_no_implicit_post_retry(self):
        import urllib.error
        with patch.object(m.urllib.request, 'urlopen', side_effect=urllib.error.HTTPError('https://api.openai.com', 429, 'rate limit', {}, None)) as request:
            with self.assertRaisesRegex(m.Failure, 'HTTP 429'):
                m.http('https://api.openai.com/v1/agents/sessions', method='POST', body={'environment': {'type': 'self_hosted'}}, key='test-key', beta=True)
        self.assertEqual(request.call_count, 1)
        sent = request.call_args.args[0]
        self.assertEqual(sent.headers['Openai-beta'], 'agents=v1')
        self.assertEqual(sent.headers['Authorization'], 'Bearer test-key')
        self.assertEqual(sent.method, 'POST')

    def test_cleanup_id_persisted_before_executor_starts(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / 'session.json'
            agent = m.Agents({'model': 'example'}, {'run': '123', 'session_file': str(path)})
            agent.call = lambda *args, **kwargs: {'id': 'sess_123'}
            agent.__enter__()
            self.assertEqual(json.loads(path.read_text()), {'id': 'sess_123'})
            agent.__exit__(None, None, None)
            self.assertFalse(path.exists())

    def test_workspace_limit_and_failure_remove_container(self):
        config = {'image': 'test-image', 'command_seconds': 1}
        calls = []
        def runner(args, **kwargs):
            calls.append(args)
            if args[1] == 'cp':
                raise m.Failure('copy failed')
            return b''
        with patch.object(m, 'command', side_effect=runner):
            with self.assertRaises(m.Failure):
                with m.Sandbox(config, fixture()):
                    pass
        self.assertEqual(calls[-1][1:4], ['rm', '-f', '-v'])


class OrchestrationTests(unittest.TestCase):
    def test_repair_budget_exhaustion_never_writes_patch(self):
        from contextlib import ExitStack
        config = {'max_archive_bytes': 100000, 'max_patch_bytes': 100000, 'max_repairs': 2}
        base = fixture()
        candidate = copy.deepcopy(base)
        candidate[m.MANIFEST] = (base[m.MANIFEST][0].replace(b'"1.0"', b'"1.0.1"'), 0o644)
        class Box:
            def __init__(self, *args): pass
            def __enter__(self): return self
            def __exit__(self, *args): pass
            def run(self, *args): return ''
            def snapshot(self): return candidate
        class Agent(Box):
            repairs = 0
            closed = False
            def connect(self, *args): pass
            def repair(self, *args): self.__class__.repairs += 1
            def __exit__(self, *args): self.__class__.closed = True
        with tempfile.TemporaryDirectory() as tmp, ExitStack() as stack:
            args = type('Args', (), {'mode': 'dry-run', 'kind': 'dependencies', 'output': Path(tmp)})()
            stack.enter_context(patch.dict(os.environ, {'GITHUB_ACTIONS': 'true', 'OPENAI_API_KEY': 'app', 'OPENAI_EXECUTOR_API_KEY': 'executor'}))
            stack.enter_context(patch.object(m, 'command', side_effect=[b'abc', m.pack(base)]))
            stack.enter_context(patch.object(m, 'existing_pr', return_value=[]))
            stack.enter_context(patch.object(m, 'github_get', return_value=[]))
            stack.enter_context(patch.object(m, 'detect', return_value=(candidate, ['serde update'])))
            stack.enter_context(patch.object(m, 'verify', return_value='build failed'))
            stack.enter_context(patch.object(m, 'Sandbox', Box))
            stack.enter_context(patch.object(m, 'Agents', Agent))
            with self.assertRaisesRegex(m.Failure, 'budget exhausted'):
                m.execute(args, config, {})
            self.assertEqual(Agent.repairs, 2)
            self.assertTrue(Agent.closed)
            self.assertFalse((Path(tmp) / 'update.patch').exists())

    def test_publisher_rejects_tampered_artifact_before_remote_calls(self):
        import publish
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'report.json').write_text(json.dumps({'status': 'verified', 'mode': 'publish', 'run': '123', 'patch_sha256': 'bad'}))
            (root / 'update.patch').write_bytes(b'tampered')
            with patch.dict(os.environ, {'GITHUB_ACTIONS': 'true', 'GITHUB_REPOSITORY': 'wata-gh/puyo2', 'GITHUB_RUN_ID': '123'}), patch.object(sys, 'argv', ['publish.py', '--input', tmp]), patch.object(publish, 'github_get') as gh:
                with self.assertRaisesRegex(m.Failure, 'digest mismatch'):
                    publish.main()
                gh.assert_not_called()


if __name__ == '__main__':
    unittest.main()
