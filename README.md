# puyo2

ぷよぷよ通の解析用基本処理です。

# Goバージョン運用方針

このリポジトリでは、Goの言語バージョンとツールチェーンを次の方針で運用します。

- **追従方針**: Goは**常に最新安定版へ追従**する。
  - 目安として、Goの安定版リリース後に内容を確認し、問題なければ速やかに更新する。
- **単一の正本 (SSOT)**: バージョンの正本は `go.mod` とし、`go` と `toolchain` を更新する。
  - CI (`.github/workflows/go.yml`) は `actions/setup-go` の `go-version-file: go.mod` を利用し、`go.mod` と常に整合させる。

## 依存更新PRの分離ルール

更新理由の切り分けをしやすくするため、依存更新は次の単位でPRを分けます。

1. **ライブラリのみ更新PR**
   - 例: `go get -u` や `go mod tidy` による `require` / `go.sum` の更新。
   - `go` / `toolchain` は変更しない。
2. **Goツールチェーン更新PR**
   - `go.mod` の `go` / `toolchain` と、それに伴うCI確認のみを扱う。
   - 原則としてライブラリ更新を混ぜない。

## 更新手順（推奨）

1. 目的に応じて「ライブラリ更新」か「Goツールチェーン更新」かを決める。
2. 変更対象を限定して実施する（混在させない）。
3. `go test ./...` で検証する。
4. PR本文に、更新種別（ライブラリ/ツールチェーン）と検証結果を明記する。

# 使い方

```
package main

import (
	"flag"
	"fmt"

	"github.com/wata-gh/puyo2"
)

func main() {
	param := flag.String("param", "a78", "puyofu")
	out := flag.String("out", "", "output file path")
	flag.Parse()
	bf := puyo2.NewBitFieldWithMattulwan(*param)
	bf.ShowDebug()
	if *out == "" {
		*out = *param + ".png"
	}
	bf.ExportImage(*out)
	fmt.Println(bf.MattulwanEditorUrl())
}
```

## IPS なぞぷよ URL 変換

`ips.karou.jp/simu/pn.html?...` のクエリを `puyo2` 向けの `Initial Field` と `Haipuyo` に変換できます。

### URL を渡す

```bash
go run ./cmd/pnconv -url 'https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270'
```

### 生クエリを渡す

```bash
go run ./cmd/pnconv -param '80080080oM0oM098_4141__u03'
```

### 出力例

```text
Initial Field: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaececeacecececececececece
Haipuyo: gbgbgb
Condition: 6連鎖する
ConditionCode: q0=30 q1=0 q2=6
```

`jjgqqqqqqqqq_q1q1q1__u06`（問2）は URL 実装準拠で `...ecece...` 系の盤面として解釈されます。

## IPS なぞぷよ Solve(JSON)

`pnconv` が変換専用なのに対して、`pnsolve` は実際に探索して条件一致解を JSON で出力します。

### 実行例

```bash
go run ./cmd/pnsolve -param '800F08J08A0EB_8161__270'
go run ./cmd/pnsolve -url 'http://ips.karou.jp/simu/pn.html?jjgqqqqqqqqq_q1q1q1__u06'
```

### 主なオプション

- `-url`: IPS なぞぷよ URL（`-param` より優先）
- `-param`: クエリ文字列
- `-disablechigiri`: ちぎり禁止で探索
- `-pretty=false`: 1行 JSON 出力

### JSON フィールド

- `input`: 入力文字列
- `initialField`: 初期盤面(Mattulwan 78文字)
- `haipuyo`: はいぷよ列(`rgbyp`)
- `condition`: `q0/q1/q2/text`
- `status`: `ok` / `no_solution` / `search_failed`
- `error`: `status=search_failed` のときのエラー内容
- `searched`: 探索した手順数
- `matched`: 条件一致した解数
- `solutions[]`: 各解（`hands/chains/score/clear/initialField/finalField`）
  - `clear` は「条件一致」ではなく「`finalField` が全消し（オールクリア）か」を表します

## pnsolve2simus

`cmd/pnsolve/pnsolve2simus` は `cmd/pnsolve` の JSON 出力を受け取り、`puyo-rsrch-app` の `/simus` URL に変換する Ruby スクリプトです。

### 使い方

`pnsolve` の結果を標準入力から渡す:

```bash
go run ./cmd/pnsolve -param '800F08J08A0EB_8161__270' -pretty=false | ./cmd/pnsolve/pnsolve2simus
```

JSON ファイルを引数で渡す:

```bash
./cmd/pnsolve/pnsolve2simus result.json
```

ローカル開発用 URL を出す（`http://localhost:3000`）:

```bash
./cmd/pnsolve/pnsolve2simus --local result.json
```

### 入出力

- 入力: `pnsolve` の JSON（トップレベル object、`initialField` と `solutions[].hands` が必要）
- 出力: 解ごとに1行の URL（`https://puyo-rsrch.com/simus?fs=...&h=...`）
