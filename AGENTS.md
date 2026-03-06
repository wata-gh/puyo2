# Repository Guidelines

## Project Structure & Module Organization
- The workspace root is a virtual Cargo workspace. The actual package is `crates/puyo2`.
- Core library code lives under `crates/puyo2/src/`:
  - engine and search modules in `*.rs`
  - CLI binaries in `crates/puyo2/src/bin/*.rs`
- Integration tests live under `crates/puyo2/tests/`.
- CI is defined in `.github/workflows/rust.yml`.

## Build, Test, and Development Commands
- `cargo build --workspace --all-targets` builds the library and all binaries.
- `cargo test --workspace` runs the full Rust test suite.
- `cargo test --workspace search_v2_opt` runs a focused subset while iterating.
- `cargo run -p puyo2 --bin nazo -- -param a78 -hands rgby` runs the search CLI.
- `cargo run -p puyo2 --bin dfield -- -param a78 -out field.png` exports a field image.

## Coding Style & Naming Conventions
- Use standard Rust style and run `cargo fmt --all`.
- Keep library code in the `puyo2` crate; binaries stay under `src/bin`.
- File names follow `snake_case.rs`.
- Exported types use `PascalCase`; functions and modules use `snake_case`.
- Keep functions focused and avoid mixing search, rendering, and parsing concerns in one function.

## Testing Guidelines
- Add integration tests under `crates/puyo2/tests/`.
- Prefer small unit tests near implementation when the logic is local and isolated.
- Prefer deterministic Mattulwan field strings in test fixtures.
- For bug fixes, add a regression test that fails before the fix and passes after it.
- CI currently enforces passing `cargo test --workspace`, so maintain or improve coverage in touched code.

## Commit & Pull Request Guidelines
- Match existing history: short imperative commit subjects like `Fix bug`, `Add utils`, or `Refactor`.
- Keep commits scoped; separate refactors from behavior changes.
- PRs should target `main`.
- Include what changed and why.
- Include validation steps (for example `cargo test --workspace`).
- Link issue(s) when available.
- Add sample output or an image when rendering/output behavior changes.

## Configuration Tips
- Image export expects sprite files such as `puyos.png` (and shape exports may use `puyos_shape.png`).
- Set `PUYO2_CONFIG` to a directory containing these assets, or provide an `images/` directory in the project root.
