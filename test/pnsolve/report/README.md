# pnsolve HTML Report

`test/pnsolve/report` は `test/pnsolve/check` の比較結果 artifact を読み込み、Bun で HTML レポートを配信するためのディレクトリです。

## できること

- ダッシュボードで各 level の実行状況を確認
- level ごとの差分件数、error 件数、`matched` の増減を一覧表示
- 各問題の `condition.text`、オリジナルページ、solution リンクを表示
- baseline JSON と current JSON の左右 diff を表示

## 前提

- `bun` が使えること
- 表示用 artifact が先に生成されていること
- artifact は `test/pnsolve/check -out-dir ...` で作成する

## 使い方

artifact を生成します。

```bash
test/pnsolve/check -out-dir test/pnsolve/artifacts/latest level1 level2
```

レポートサーバを起動します。

```bash
cd test/pnsolve/report
bun run start -- --artifact ../artifacts/latest --port 8787
```

ブラウザで開きます。

```text
http://127.0.0.1:8787/
```

## 開発用コマンド

監視付き起動:

```bash
bun run dev -- --artifact ../artifacts/latest --port 8787
```

テスト:

```bash
bun test
```

## 画面構成

- `/`
  - ダッシュボード
  - 全 level の `実行済み / 未実行` と差分サマリを表示
- `/levels/:level`
  - level 詳細
  - 各問題の一覧、リンク、差分、raw JSON を表示

## オプション

- `--artifact <DIR>`
  - 読み込む artifact ディレクトリを指定
  - 省略時は `test/pnsolve/artifacts/latest`
- `--port <PORT>`
  - HTTP ポートを指定
  - 省略時は `8787`

## 補足

- このサーバは `pnsolve` を再実行しません
- 差分判定は `test/pnsolve/check` と同じ正規化ルールに従います
- solution リンクは `http://localhost:3000/simus?...` を使います
