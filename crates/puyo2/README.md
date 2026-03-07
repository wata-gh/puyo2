# puyo2 crate

Install a specific CLI from git:

```bash
cargo install --git https://github.com/wata-gh/puyo2 --locked -p puyo2 --bin pnsolve
```

Available binaries:

- `dfield`
- `dpuyo`
- `nazo`
- `pnconv`
- `pnsolve`
- `pnsolve2simus`

Image-oriented binaries (`dfield`, `dpuyo`) require external sprite assets.
Set `PUYO2_CONFIG` to a directory that contains `puyos.png`,
`puyos_transparent.png`, and `puyos_shape.png`, or provide an `images/`
directory in the current working directory.

The repository root `README.md` contains full build, test, and usage notes.
