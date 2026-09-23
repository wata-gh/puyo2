#!/bin/bash
set -euo pipefail
# This script always comes from the trusted base, never from agent output.
run() { printf "CHECK: "; printf "%q " "$@"; printf "\n"; "$@"; }
run cargo fmt --all --check
run cargo build --workspace --all-targets --locked
run cargo test --workspace --locked
MSRV=$(python3 -c 'import tomllib; print(tomllib.load(open("crates/puyo2/Cargo.toml", "rb"))["package"]["rust-version"])')
run rustup toolchain install "$MSRV" --profile minimal
run cargo +"$MSRV" build --workspace --all-targets --locked
run cargo +"$MSRV" test --workspace --locked
run cargo build --workspace --release --bins --locked
run test/pnsolve/check -bin "${CARGO_TARGET_DIR:-target}/release/pnsolve" level1 level2 level3 level4 level5
run cargo package --manifest-path crates/puyo2/Cargo.toml --allow-dirty --locked
run cargo install --path crates/puyo2 --locked --bin pnsolve --root /tmp/puyo2-install
run /tmp/puyo2-install/bin/pnsolve -param '800F08J08A0EB_8161__270' -pretty=false >/dev/null
