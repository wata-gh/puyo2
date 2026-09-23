# puyo2 の定期更新

`Maintenance updates` は puyo2 専用の更新 PR 作成処理です。マージ、リリース公開、他リポジトリの変更はしません。初期状態では定期実行は無効です。

## 処理と信頼境界

1. `main` のコミットを固定し、同種の開いている更新 PR を確認します。
2. ライブラリ更新では crates.io の安定版・非 yanked・宣言 MSRV 以下の直接依存を選び、バージョン指定を更新して `cargo update` します。既知の major 更新も候補に含みます。MSRV 未宣言の依存は実際の MSRV build/test で検証します。Rust 更新では公式 stable manifest の実行版だけを更新します。
3. Git 履歴・認証情報を含めないソースのスナップショットを、毎回新しい Docker コンテナに渡します。ホストのディレクトリ、Docker socket、GitHub token、OpenAI アプリ用キーはマウント・注入しません。非 root、capabilities 全削除、PID 256、メモリ 6 GiB、CPU 2 の制限があります。
4. Cargo 解決や検証が失敗した場合だけ Agents API の `self_hosted` セッションを作ります。API はホスト上の Python コントローラから呼び出し、コンテナ内の `codex exec-server` に渡すのは環境接続専用キーだけです。不要なモデル呼び出しは行いません。
5. 回収したアーカイブをデータとして読み、パストラバーサル・リンク・サイズ超過・ファイル削除・モード変更を拒否します。変更可能なのは依存 version、Cargo.lock、Rust 更新時の toolchain channel、必要な Rust ソース、新しい回帰テストだけです。既存 integration test・fixture・inline test、MSRV、edition、package metadata、依存 feature/source、workflow、検証コードは保護します。ソース修正には新規回帰テストを要求します。
6. エージェント環境とは別の新規コンテナで、信頼した `verify.sh` を実行します。エージェントの完了状態・報告・生成したログだけでは成功にしません。
7. 全検証に成功した差分と結果を artifact に保存します。独立した publish job が基準 SHA、差分ハッシュ、ポリシーを再検査し、GitHub App token で通常の draft PR を作成します。候補コードは publish job で実行しません。通常の `Rust` PR workflow が公開した正確な HEAD で成功するまで最大 30 分確認します。失敗・未起動・時間切れは失敗として記録し、PR は draft のままです。

Agents API のアプリケーションキーは実行環境の外に置くのが公式の構成です。[Self-hosted sandboxes](https://developers.openai.com/api/docs/guides/agents-api/environments/self-hosted)

初期範囲では `rust-version = 1.94` と `edition = 2024` を自動変更しません。これらを上げなければ成立しない更新は、人間が別 PR で判断します。依存の最大 version より MSRV 維持を優先します。意味的な無関係変更や公開 API 破壊を完全に機械判定するものではありません。保護ゲート・既存テスト・新規回帰テストに加え、人間の diff review が必須です。自動生成テストの妥当性もレビューしてください。

## 検証内容

- `cargo fmt --all --check`
- 固定実行版の `cargo build --workspace --all-targets --locked` と `cargo test --workspace --locked`
- 宣言 MSRV で同じ build/test
- release の全バイナリ build
- 既存 `test/pnsolve/check` の level1～5（正規化済み baseline 比較）
- `cargo package` のパッケージ内再ビルド、`cargo install --locked --bin pnsolve`、インストール後の CLI smoke
- 通常 PR CI: Linux、Intel macOS、Apple Silicon macOS の build/test/release build/package/install、Linux MSRV、コントローラの失敗系テスト

画像出力を含む既存 integration tests を利用します。release workflow も repository の固定 Rust と lockfile を利用するよう揃えています。今回 release は実行しません。

2026-09-23 に確認した `puyo-rsrch-engine` の workspace は `puyo2 = { git = "https://github.com/wata-gh/puyo2.git", branch = "main", package = "puyo2" }` を指定していました。下流の lock 更新時には新しい main を取り込む可能性があります。この処理では下流の build/test や lock 更新を行いません。公開型・関数・feature と、依存 version の共存、MSRV をレビューしてください。下流検証は別途必要です。

## 有効化

1. この実装をレビューして main にマージします。本タスクではマージしません。
2. OpenAI Platform で専用 project/service account と、選んだ model（設定の初期値 `gpt-6-astra`）へのアクセスを確認します。Agents API は beta です。アプリ用キーには `api.agents.read`、`api.agents.write`、`api.responses.write` を許可します。vault は使いません。
3. 同じ organization/project/service account の Agents タブで環境接続キーを作成し、それ以外の権限を None にします。
4. GitHub Actions secrets に、秘密管理から直接以下を登録します。ローカル `.env`、mise 設定、ファイル、shell history、ログには平文で保存しません。`set -x` は使用しません。

| 種類 | 名前 | 内容 |
| --- | --- | --- |
| Secret | `MAINTENANCE_OPENAI_API_KEY` | ホスト側だけのアプリ用キー |
| Secret | `MAINTENANCE_EXECUTOR_API_KEY` | コンテナ内 `CODEX_API_KEY` に渡す環境接続キー |
| Secret | `MAINTENANCE_APP_PRIVATE_KEY` | puyo2 専用 GitHub App の秘密鍵 |
| Variable | `MAINTENANCE_APP_ID` | 同 App の ID |
| Variable | `MAINTENANCE_ENABLED` | 定期公開を有効にする場合のみ `true` |

GitHub App は puyo2 のみへインストールし、repository permissions を Contents: write、Pull requests: write、Actions: read、暗黙の Metadata: read に制限します。Actions の更新権限、Administration、Secrets、Checks write は不要です。branch protection の bypass は与えません。作成 token は publish job にだけ存在し、action の終了処理で失効します。

`GITHUB_TOKEN` で作成・更新した PR の CI は承認待ちになるため、通常 CI を自動起動する公開には App token を使います（[GitHub のトリガー仕様](https://docs.github.com/en/actions/how-tos/writing-workflows/choosing-when-your-workflow-runs/triggering-a-workflow)）。update job の `GITHUB_TOKEN` は Contents/Pull requests read で API の読み取りだけに使います。GitHub の Actions/Apps 利用ポリシーが App 作成 PR の通常 CI を許可することを確認してください。`pull_request_target` で候補コードを実行する仕組みは使いません。

5. まず手動 `dry-run` を実行し、artifact の `report.json` と `update.patch` をレビューします。キー未設定でも、API を必要としない検出・決定的更新・検証は動きます。互換性修正が必要な段階でキー不足なら失敗します。
6. `publish` で初回の draft PR を作り、通常 CI の全 matrix が実行・成功したこと、App の権限、費用、cleanup を確認します。その後に `MAINTENANCE_ENABLED=true` とします。サーバー側 API アクセスと実セッションの動作はこの導入確認が必要です。

## スケジュールと手動実行

- 依存: 月曜 02:17 UTC（11:17 JST）、週次
- Rust: 毎月 1 日 02:43 UTC（11:43 JST）、月次
- GitHub の schedule は遅延・間引きがあり、時刻保証ではありません。
- 同時実行 group は両カテゴリ共通です。実行中 job をキャンセルしません。GitHub concurrency の pending は置き換わる場合があります。
- 同種の PR が開いていればスキップします。既存 PR を自動で上書きしません。
- ブランチは `codex/maintenance-dependencies` と `codex/maintenance-rust`。main が検証中に進んだ場合は公開せず再実行を要求します。push は存在しない ref を条件にする lease を指定し、確認後の競合でも既存 branch を上書きしません。
- PR の merge/close 後は更新ブランチを削除します（repository の自動削除設定を推奨）。PR なしの残存ブランチは安全のため失敗し、人間の確認を要求します。

```sh
gh workflow run maintenance.yml --ref main -f kind=dependencies -f mode=dry-run
gh workflow run maintenance.yml --ref main -f kind=rust -f mode=dry-run
gh workflow run maintenance.yml --ref main -f kind=dependencies -f mode=publish
gh workflow run maintenance.yml --ref main -f kind=dependencies -f mode=offline
```

`offline` は候補検出・OpenAI/GitHub API・公開を行わず、コミット済みソースの実環境検証をします。Docker build、Rust/Cargo パッケージ取得にはネットワークが必要です。「完全なネットワーク遮断」ではありません。`dry-run` は候補検出と必要時の実 Agents API 修正まで行うため、費用が発生し得ますが branch/PR は作りません。

macOS + mise で秘密値なしのローカル検証:

```sh
mise exec -- python3 -m unittest discover -s ops/maintenance/tests -v
docker build -t puyo2-maintenance:local ops/maintenance
# コミット済み HEAD が入力です。ローカル未コミット変更は渡しません。
mise exec -- python3 ops/maintenance/maintenance.py --mode offline --output /tmp/puyo2-maintenance-output
```

Python 3.11 以上、Docker、Git が必要です。検証ツール（Rust、zsh、jq、jd、Codex）は image に入ります。live モードは使い捨て GitHub Actions runner に限定し、ローカル Docker metadata に環境キーを残す運用を避けます。mise の secret 設定は追加していません。

## 上限と費用

`ops/maintenance/config.json` のレビューで変更します。workflow input から任意コマンド・モデル・上限・image を渡すことはできません。

| 制御 | 初期値 |
| --- | --- |
| 互換性修正 | 最大 2 turn、1 session、subagent 無効 |
| 1 turn | 600 秒 |
| Controller 全体 | 2,400 秒（cleanup 用の余裕あり） |
| コマンド | 900 秒 |
| 観測 token | 100,000（入力 + 出力、session の全 turn） |
| usage が不明 | 60 秒で中断、成功扱いしない |
| 回収アーカイブ | 16 MiB |
| 公開 patch | 512 KiB |
| update job / publish job | 55 分 / 35 分 |
| 通常 PR CI の確認待ち | 最大 30 分 |
| artifact 保持 | 7 日 |

usage は best effort で、遅れて増えることがあります。これは **観測値での中断閾値であり、厳密な token/金額の課金上限ではありません**。公式仕様で未確認の `max_tokens` 等を送って上限保証とはしません。専用 project の利用制限・通知も設定し、利用量監視とキー失効を運用に含めてください。ダッシュボードの予算アラートが強制停止とは限りません。価格はモデル・キャッシュ等に依存するため固定見積もりをコードに入れていません。[Observability and usage](https://developers.openai.com/api/docs/guides/agents-api/observability)

## 失敗・停止・後始末

- `report.json`: 候補、基準 SHA、実コマンドの成功/失敗と出力末尾、使用量、修正回数、失敗理由、下流未検証を保存。キーと既知 token 形式は保存前に伏せます。生の executor ログ・API レスポンスは保存しません。
- テスト不一致、fmt 不一致、配布失敗: immutable baseline と diff を確認します。既存テストを書き換えて通しません。修正 2 回で終わらなければ人間に引き継ぎます。
- 401/403: App/API の権限・project・model access・環境キーの所有者一致を確認。ほかの認証経路に自動で切り替えません。
- 429/通信切断/POST 応答喪失: エラーとして終了。作成・入力 POST の自動リトライはしません（重複 session/turn 防止）。作成がサーバーで成立して応答だけ失われた場合は session ID が取得できないことがあるため、Agents ダッシュボードで metadata `purpose=puyo2-maintenance`, `run=<GitHub run ID>` の残存 session を確認します。
- usage 未確定: 0 と見なさず停止します。API 側が turn 終了まで usage を返さない場合もこの厳格な watchdog で停止し得ます。導入 dry-run で実挙動を確認し、必要なら上限ポリシーをレビューしてください。
- 後始末: 通常終了・失敗・SIGTERM・全体 deadline で turn cancel と session DELETE を実行し、コンテナは `docker rm -f -v`。さらに Actions の `always()` step が外側に保存した session ID と label 付きコンテナを回収します。削除失敗は黙殺せず job を失敗させます。runner 強制喪失時は dashboard で session を削除してください（コンテナは runner と共に廃棄）。物理的な API データ削除は非同期になり得ます。
- 公開中断: branch ができて PR がない場合は diff と commit を確認し、人間が PR を作るか branch を削除して再実行します。再実行で branch を force 更新しません。
- 停止: `MAINTENANCE_ENABLED` を削除または `false` にします。手動実行も止める場合は `gh workflow disable maintenance.yml`。実行中 run をキャンセルし cleanup を確認、緊急時は OpenAI 両キーと App を失効します。すでに開いた PR は自動で merge されません。

## 仕様確認と未検証事項

2026-09-23 に OpenAI Docs の [overview](https://developers.openai.com/api/docs/guides/agents-api/overview)、[quickstart](https://developers.openai.com/api/docs/guides/agents-api/quickstart)、[self-hosted](https://developers.openai.com/api/docs/guides/agents-api/environments/self-hosted)、[OpenAI-hosted](https://developers.openai.com/api/docs/guides/agents-api/environments/openai-hosted)、session/turn/usage の公式仕様を確認しました。Rust と既存 CLI harness を固定でき、同じソースを独立検証できる self-hosted container を採用しました。

公開 npm `openai@7.22.0` に `beta.agents.sessions.create` が存在し、`@openai/codex@0.157.0-alpha.11` に `exec-server --remote --environment-id` が存在することを認証なしで確認しました。コントローラは Python 標準ライブラリから公式 REST `/v1/agents/sessions` に `OpenAI-Beta: agents=v1` を付けてアクセスします。旧 Agents SDK / Responses API に読み替えていません。SDK 依存は不要です。

実 API キーを設定した session 接続・修正・課金・削除、および専用 GitHub App による更新 PR の作成は未検証です。オフラインの契約/失敗系テストはアカウントの API 利用権限を保証しません。有効化前に上記の dry-run/publish 導入確認を完了してください。

実装時の検証: macOS の `cargo test --workspace --locked`、24 件のコントローラ契約・失敗系テスト、actionlint 1.7.12 が成功しました。使い捨て Linux ARM64 コンテナで固定版/MSRV build・test、release build、pnsolve level1～5（247 件、diff/missing/run error/jd error すべて 0）、package 再ビルド、install と CLI smoke が成功しました。コンテナ snapshot の往復一致と cleanup も確認しています。
