# puyo2

ぷよぷよ通の解析用基本処理です。

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
