#!/usr/bin/env python3
"""Idempotent fallback on the disposable Actions runner; no logs from executor."""
import argparse
import json
import os
from pathlib import Path
from maintenance import command, http, Failure

parser = argparse.ArgumentParser()
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
failed = False
# Session ID is persisted immediately after creation, outside the sandbox.
p = args.output / 'session.json'
if p.exists():
    session = json.loads(p.read_text())['id']
    try:
        http('https://api.openai.com/v1/agents/sessions/' + session, method='DELETE',
             key=os.environ['OPENAI_API_KEY'], beta=True)
        p.unlink()
    except Failure as e:
        if 'HTTP 404:' in str(e):
            p.unlink()
        else:
            print('Session cleanup failed; use saved session ID to delete it. ' + str(e))
            failed = True
ids = command(['docker', 'ps', '-aq', '--filter', 'label=puyo2-maintenance=true']).decode().split()
if ids:
    command(['docker', 'rm', '-f', '-v', *ids])
raise SystemExit(1 if failed else 0)
