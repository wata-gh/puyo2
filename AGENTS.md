# Repository Guidelines

## Project Structure & Module Organization
- The root `puyo2` package holds the core engine: bit-field operations (`bit_field*.go`), search logic (`search*.go`, `search_v2.go`), shape processing (`shape_bit_field*.go`), scoring (`score.go`), and shared helpers.
- Tests are colocated with source files as `*_test.go`.
- Command-line tools live under `cmd/`.
- `cmd/dfield` renders field images.
- `cmd/dpuyo` draws sprite assets.
- `cmd/nazo` runs puzzle/search exploration.
- CI is defined in `.github/workflows/go.yml`.

## Build, Test, and Development Commands
- `go build ./...` builds the library and all commands.
- `go test ./...` runs the full test suite.
- `go test -run TestPlacePuyo ./...` runs a focused test while iterating.
- `go test -bench . ./...` runs benchmarks for hot paths.
- `go run ./cmd/nazo -param a78 -hands rgby` runs the search CLI.
- `go run ./cmd/dfield -param a78 -out field.png` exports a field image.

## Coding Style & Naming Conventions
- Use standard Go style and run `gofmt` on changed files.
- Keep core code in `package puyo2`; command entry points use `package main`.
- File names follow `snake_case.go` (for example, `shape_bit_field_export.go`).
- Exported identifiers use `PascalCase`; internal helpers use `camelCase`.
- Keep functions focused and avoid mixing search, rendering, and parsing concerns in one function.

## Testing Guidelines
- Add tests next to implementation in `*_test.go` with `TestXxx`/`BenchmarkXxx` names.
- Prefer deterministic Mattulwan field strings in test fixtures.
- For bug fixes, add a regression test that fails before the fix and passes after it.
- CI currently enforces passing `go test ./...` (no explicit coverage threshold), so maintain or improve coverage in touched code.

## Commit & Pull Request Guidelines
- Match existing history: short imperative commit subjects like `Fix bug`, `Add utils`, or `Refactor`.
- Keep commits scoped; separate refactors from behavior changes.
- PRs should target `main`.
- Include what changed and why.
- Include validation steps (for example `go test ./...`).
- Link issue(s) when available.
- Add sample output or an image when rendering/output behavior changes.

## Configuration Tips
- Image export expects sprite files such as `puyos.png` (and shape exports may use `puyos_shape.png`).
- Set `PUYO2_CONFIG` to a directory containing these assets, or provide an `images/` directory in the project root.
