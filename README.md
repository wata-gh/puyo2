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
