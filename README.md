# puyo2

Rust で実装した、ぷよぷよ通の解析ライブラリと CLI 群です。

## Project Layout

- `crates/puyo2/src/lib.rs`: 公開 library crate `puyo2`
- `crates/puyo2/src/bin/*.rs`: CLI binaries
  - `dfield`
  - `dpuyo`
  - `nazo`
  - `pnconv`
  - `pnsolve`
  - `pnsolve2simus`
- `crates/puyo2/tests/*.rs`: library と CLI の integration tests
- `.github/workflows/rust.yml`: Rust build/test CI

## Build And Test

```bash
cargo build --workspace --all-targets
cargo test --workspace
```

## Example Commands

```bash
cargo run -p puyo2 --bin pnconv -- -url 'https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270'
cargo run -p puyo2 --bin pnsolve -- -param '800F08J08A0EB_8161__270'
cargo run -p puyo2 --bin pnsolve2simus -- result.json
cargo run -p puyo2 --bin nazo -- -param a78 -hands rgby
cargo run -p puyo2 --bin dfield -- -param a78 -out field.png
```

## Assets

画像出力は `puyos.png`、`puyos_transparent.png`、`puyos_shape.png` を使います。

- `PUYO2_CONFIG` が設定されていれば、そのディレクトリを優先して参照します
- 未設定の場合は repo 直下の `images/` を参照します

## pnconv

`ips.karou.jp/simu/pn.html?...` のクエリを `Initial Field` と `Haipuyo` に変換します。

```bash
cargo run -p puyo2 --bin pnconv -- -param '80080080oM0oM098_4141__u03'
```

出力例:

```text
Initial Field: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaececeacecececececececece
Haipuyo: gbgbgb
Condition: 6連鎖する
ConditionCode: q0=30 q1=0 q2=6
```

`jjgqqqqqqqqq_q1q1q1__u06`（問2）は URL 実装準拠で `...ecece...` 系の盤面として解釈されます。

## pnsolve

`pnsolve` は条件一致解を JSON で出力します。

```bash
cargo run -p puyo2 --bin pnsolve -- -param '800F08J08A0EB_8161__270'
cargo run -p puyo2 --bin pnsolve -- -url 'http://ips.karou.jp/simu/pn.html?jjgqqqqqqqqq_q1q1q1__u06'
```

主なオプション:

- `-url`: IPS なぞぷよ URL（`-param` より優先）
- `-param`: クエリ文字列
- `-disablechigiri`: ちぎり禁止で探索
- `-pretty=false`: 1行 JSON 出力
- `-expand-equivalent-hands`: 停止連鎖ベースの exhaustive expansion を行います
  - 内部では `stop_on_chain=true` と `dedup=off` を使います

JSON フィールド:

- `input`: 入力文字列
- `initialField`: 初期盤面（Mattulwan 78 文字）
- `haipuyo`: はいぷよ列（`rgbyp`）
- `condition`: `q0/q1/q2/text`
- `status`: `ok` / `no_solution` / `search_failed`
- `error`: `status=search_failed` のときのエラー内容
- `searched`: 探索した手順数
- `matched`: 条件一致した解数
- `solutions[]`: 各解（`hands/chains/score/clear/initialField/finalField`）
  - `clear` は `finalField` が全消しかどうかを表します

## pnsolve2simus

`pnsolve2simus` は `pnsolve` の JSON 出力を `/simus` URL に変換します。

```bash
cargo run -p puyo2 --bin pnsolve -- -param '800F08J08A0EB_8161__270' -pretty=false \
  | cargo run -p puyo2 --bin pnsolve2simus --
```

```bash
cargo run -p puyo2 --bin pnsolve2simus -- --local result.json
```

入力は `pnsolve` の JSON object、出力は解ごとに 1 行の URL です。
