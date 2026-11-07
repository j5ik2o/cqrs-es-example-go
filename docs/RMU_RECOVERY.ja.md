# Read Model Updater の復旧

この文書は、Read Model Updater（RMU）が失敗したときに、再試行・復旧タスク・再同期でどう立て直すかを定める。対象は `pkg/rmu`、Lambda 版の `rmu` コマンド、ローカル版の `localRmu` コマンド、復旧用の `recoverRmu` コマンドと `recoverRmuWorker` コマンド、デプロイ資産の `deploy/rmu` である。

## 1. 失敗を二つに分ける

RMU は失敗を、保存先全体の障害（transient）と、特定のイベントが何度試しても通らない失敗（permanent）に分ける。この区別が再試行と復旧タスクの分かれ目になる。

- transient は MySQL の接続断・デッドロック・ロック待ちタイムアウト、DynamoDB のスロットリング、ネットワーク断である。回数と経過時間の上限まで再試行し、上限に達したら処理を止める。
- permanent はイベント本文の JSON 不正、未知のイベント型、journal の欠番、リードモデルの制約違反である。同じ入力を繰り返しても直らないので、再試行せずに復旧タスクを記録する。

どちらの場合も、未適用のイベントがある集約は、その位置より先へ進めない。位置 `p` のイベントが未適用なら、`p+1` 以降は適用しない。この規則はストリーム処理、再同期、復旧のすべてで同じ `Applier` が守る。

## 2. 再試行

Lambda 版の再試行の上限は、既定で試行 5 回、経過時間 30 秒である。待ち時間は 200ms から始まり、5 秒で頭打ちにしたうえで full jitter をかける。ローカル版は 500ms から始まり、30 秒で頭打ちにする。設定は環境変数で変えられる。

| 環境変数 | 既定値 | 意味 |
|:--|:--|:--|
| `RMU_RETRY_MAX_ATTEMPTS` | `5` | Lambda 版の最大試行回数 |
| `RMU_RETRY_MAX_ELAPSED` | `30s` | Lambda 版の最大経過時間 |
| `RMU_RETRY_INITIAL_DELAY` | `200ms` | 最初の待ち時間 |
| `RMU_RETRY_MAX_DELAY` | `5s` | 待ち時間の上限 |
| `STREAM_RETRY_MAX_ATTEMPTS` | `5` | ローカル版の最大試行回数 |
| `STREAM_RETRY_MAX_ELAPSED` | `30s` | ローカル版の最大経過時間 |
| `STREAM_RETRY_INITIAL_DELAY` | `500ms` | ローカル版の最初の待ち時間 |
| `STREAM_RETRY_MAX_DELAY` | `30s` | ローカル版の待ち時間の上限 |

ローカル版は、処理に失敗したときに 1 秒固定で無限に再試行するのをやめた。指数バックオフとジッターで間隔を空け、成功したら間隔を戻す。ストリームにレコードがないときのポーリングも同じ方式で間隔を空ける。

Lambda 版は部分バッチ応答を返す。成功したレコードは再配送させず、失敗したレコードだけを `batchItemFailures` に入れる。同じ集約の後続レコードも、先のイベントが未適用なら失敗として返す。Lambda のイベントソースマッピングには次の設定が必要である。

| 設定 | 値 | 理由 |
|:--|:--|:--|
| `BisectBatchOnFunctionError` | `true` | 失敗したレコードだけにバッチを分割し、正常なレコードの再処理を減らす |
| `FunctionResponseTypes` | `["ReportBatchItemFailures"]` | 部分バッチ応答を有効にする |
| `MaximumRetryAttempts` | `3` | 無限再試行を防ぐ |
| `MaximumRecordAgeInSeconds` | `86400` | 保持期間を超えたレコードを打ち切る |
| `DestinationConfig.OnFailure` | 失敗通知用の SQS | 安全網として失敗通知を届ける |

この表の値は `deploy/rmu/event-source-mapping.json` を正本にする。このリポジトリはデプロイ用の IaC を持たないため、設定を再現できるように JSON と適用スクリプトを置いた。`deploy/rmu` のテストが値の読み取りと Lambda の上限を検査する。適用の手順は「8. 本番適用の手順」に書く。

Lambda のタイムアウトは `60` 秒にする。再試行の経過時間の上限 `RMU_RETRY_MAX_ELAPSED`（既定 30 秒）の 2 倍を確保し、バッチの処理と終了に使える時間を残す。テストはタイムアウトが再試行の上限の 2 倍以上であることを確かめる。再試行の上限を変えるときは `deploy/rmu/event-source-mapping.json` の `retry` と `rmuFunction.timeoutSeconds` も同じ変更に含める。

## 3. DLQ と復旧タスク

失敗を検知した時点で、RMU は `read_model_recovery_tasks` に復旧タスクを書く。記録するのは、ストア識別子、投影名、投影の版、集約 ID、`seq_nr`、失敗種別、エラーメッセージである。

SQS は通知を届けるためだけに使う。メッセージにイベント本文は入れない。復旧処理は、タスクの識別子だけを受け取り、イベント本文を journal から読み直す。OnFailure による SQS への配送は安全網であり、アプリが書いた復旧タスクが正本である。

通知を受けるワーカーは `recoverRmuWorker` コマンドで動かす。1 回の呼び出しでバッチ全体に対して 1 回の復旧を実行する。メッセージ本文は読まない。再試行できる失敗では、バッチの全メッセージ ID を `batchItemFailures` に返して SQS に再配送させる。再試行で直らない失敗では、復旧タスクがすでに保存されているので通知を成功として扱う。

ワーカーの SQS イベントソースマッピングにも `FunctionResponseTypes = ["ReportBatchItemFailures"]` が必要である。設定しないと、`batchItemFailures` を返しても Lambda がメッセージを削除してしまう。

障害通知用のキューの可視性タイムアウトは 960 秒、ワーカーのタイムアウトは 900 秒にする。ワーカーの実行中にメッセージが再表示されないようにするためである。復旧は冪等なので、再表示が起きても二重適用にはならない。

復旧は冪等である。`recoverRmu` は次の順に動く。

1. head テーブルを強整合で Scan し、投影が遅れている集約を journal から補う。これは DY-15 の再同期である。
2. `pending` の復旧タスクを古い順に、上限まで読む。
3. タスクごとに現在位置を読み直す。位置がタスクの `seq_nr` 以上なら、適用せずに完了にする。
4. 位置が遅れていれば、journal を昇順に読み、`現在位置＋1` の規則で適用する。
5. 適用後の位置がタスクの `seq_nr` 以上になったことを確かめてから `completed` にする。

復旧タスクの書き込みに失敗した場合は、その失敗を呼び出し元に返す。タスクを失うと、永久に失敗するイベントを見落とすからである。通知の失敗は致命的ではない。タスクはすでに永続化されているので、ワーカーが次回読み直せる。

### 復旧タスクの件数上限と定期実行

`recoverRmu` と `recoverRmuWorker` は、1 回の実行で処理する復旧タスクの件数を `RECOVERY_MAX_TASKS_PER_PASS` で制限する。既定は 100 件で、1000 件を上限とする。0 以下の値と 1000 を超える値は、この範囲に丸める。

タスクは `created_at` の古い順に読む。上限に達したとき、残りのタスクは `pending` のまま次回に回す。その回の処理は失敗にしない。ログに `recovery task limit reached` を警告として出し、完了ログの `deferred` を 1 にする。次の定期実行か、次の SQS 通知が残りを処理する。

SQS のバッチを失敗として返さないのは、恒久的に失敗するタスクが残っていると再配送が終わらないためである。タスクは MySQL に残るので、次回の実行で再度試せる。

SQS の通知を失った場合に備えて、同じ復旧ワーカーを EventBridge で 5 分ごとに起動する。スケジュール実行は、head の再同期と `pending` タスクの再適用を 1 回行う。メッセージ本文は使わない。復旧は冪等なので、SQS の通知とスケジュールが同じ時刻に走っても二重適用にはならない。

スケジュール実行の失敗は呼び出し元に返す。SQS の通知のようにメッセージを成功として扱う相手がいないため、EventBridge の再試行と CloudWatch のアラームに失敗を残す。恒久的なデータの問題は人が直すまで残る。

## 4. 再同期

再同期は DY-15 の手順に従う。

1. Streams の読み取りを再開する。
2. 並行して head テーブルを `ConsistentRead = true` で Scan する。`LastEvaluatedKey` が返らなくなるまで読み切る。`aid = __config__` の設定項目は対象外にする。
3. 各 head の `seq_nr` と投影の位置を比べる。
4. 投影が遅れている集約だけ、journal から `位置＋1` 以降を適用する。
5. Streams と Scan の両方から届いたイベントは、位置のガードで重複として捨てる。

失敗の種類によって再同期の範囲を変える。

- `TrimmedDataAccessException` を受けたとき、または最後に処理した位置が保持期間の 24 時間より古いときは、head テーブルを全件照合する。
- `ExpiredIteratorException` はイテレーターの有効期限切れであり、レコードは消えていない。全件照合はせず、最後に処理したシーケンス番号の後ろからイテレーターを取り直す。

ローカル版は起動時に必ず全件照合する。初回起動には位置がなく、再起動は 24 時間を超える停止の後かもしれないからである。Lambda 版で保持期間を超えた場合は、OnFailure を契機に `recoverRmu` を実行して全件照合する。

## 5. トランザクション

リードモデルの更新と `read_model_positions` の更新は、同じ MySQL トランザクションで確定する。`STREAM_MAX_EVENTS_PER_TRANSACTION` で 1 回に確定するイベント数を決める。既定値は 1 であり、イベントごとに確定する。範囲をまとめる場合も、適用するのは `現在位置＋1` のイベントだけである。

競合は行ロックで制御する。位置行を `INSERT IGNORE` で用意してから `SELECT ... FOR UPDATE` でロックする。位置行がまだない状態で `FOR UPDATE` するとギャップロックになり、同じ隙間で新しい集約を同時に作るとデッドロックするためである。

journal に欠番があれば、位置を進めずに失敗とする。欠番を飛ばして後続を適用すると、リードモデルに穴が残る。

## 6. 追加したテスト

- 同時適用: 同じ head レコードを 2 つのゴルーチンで適用し、位置と行が重複しないことを確かめる。
- 位置行の同時作成: 別々の集約に対して同時にトランザクションを開き、両方の位置行ができることを確かめる。
- 途中の失敗: バッチの 2 件目が失敗したとき、1 件目の確定を残し、位置を失敗したイベントの手前で止める。
- 退避後の未更新: 再試行の上限に達したあと、位置もリードモデルも変わらないことを確かめる。
- 24 時間超の停止: 投影が遅れた head を再同期が journal から埋めることを確かめる。
- Scan の複数ページ: 1 ページ 1 件に制限した Scan で、全 head を読み切ることを確かめる。
- 復旧タスクの書き込み失敗: タスクを書けないときにエラーを返し、位置を進めないことを確かめる。
- 部分バッチ応答: 失敗したレコードと、同じ集約の後続レコードだけを `batchItemFailures` に入れることを確かめる。
- 復旧の冪等性: 同じタスクを 2 回適用しても、リードモデルが二重にならないことを確かめる。
- SQS 通知の成功、再試行可能な失敗、恒久的な失敗: 成功と恒久的な失敗では通知を成功として扱い、再試行可能な失敗ではバッチの全メッセージを再配送対象にすることを確かめる。
- 再同期の恒久失敗と復旧の一時失敗が同時にある場合: どちらか一方でも再試行できるなら、通知を再配送することを確かめる。
- 件数上限: 上限を超えるタスクがあるとき、上限までを処理し、残りを `pending` のまま次回に回すことを確かめる。
- 上限内の処理: タスク数が上限以下のとき、`deferred` が偽で、実行後に `pending` が 0 になることを確かめる。
- 失敗の計上: 再適用に失敗したタスクを `task_failures` に数え、`pending` に残すことを確かめる。
- 上限の設定値: 0 以下と 1000 を超える値が既定値と上限に丸まることを確かめる。
- スケジュール実行: EventBridge のイベントと空のペイロードで 1 回の復旧を実行し、失敗を呼び出し元に返すことを確かめる。
- 完了ログ: `pending`、`deferred`、`task_failures`、`resync_failed` がメトリクスフィルタで読める形で出ることを確かめる。
- デプロイ設定: `deploy/rmu/event-source-mapping.json` を読み、部分バッチ応答の有効化、タイムアウトと再試行の関係、Lambda と SQS の上限、定期実行の間隔を検査する。
- アラーム定義: `deploy/rmu/cloudwatch-alarms.json` を読み、タスク滞留・再試行上限超過・再同期失敗のアラームとメトリクスフィルタがそろっていることを検査する。

## 7. 設定

ローカル版の主な環境変数は次のとおりである。

| 環境変数 | 既定値 | 意味 |
|:--|:--|:--|
| `STREAM_HEAD_TABLE_NAME` | `head` | head テーブル名 |
| `STREAM_JOURNAL_TABLE_NAME` | `journal` | journal テーブル名 |
| `STREAM_MAX_ITEM_COUNT` | `100` | `GetRecords` の 1 回の上限 |
| `STREAM_MAX_EVENTS_PER_TRANSACTION` | `1` | 1 トランザクションで確定するイベント数 |
| `RECOVERY_MAX_TASKS_PER_PASS` | `100` | 起動時に処理する復旧タスク数の上限。最大 1000 |

Lambda 版の主な環境変数は次のとおりである。

| 環境変数 | 既定値 | 意味 |
|:--|:--|:--|
| `DATABASE_URL` | なし | MySQL の接続先。必須 |
| `PERSISTENCE_JOURNAL_TABLE_NAME` | `journal` | journal テーブル名 |
| `PERSISTENCE_HEAD_TABLE_NAME` | `head` | head テーブル名。ストア識別子の読み取りに使う |
| `PERSISTENCE_MAX_EVENTS_PER_TRANSACTION` | `1` | 1 トランザクションで確定するイベント数 |
| `RMU_RETRY_MAX_ATTEMPTS` | `5` | Lambda 版の最大試行回数 |
| `RMU_RETRY_MAX_ELAPSED` | `30s` | Lambda 版の最大経過時間 |
| `RMU_RETRY_INITIAL_DELAY` | `200ms` | 最初の待ち時間 |
| `RMU_RETRY_MAX_DELAY` | `5s` | 待ち時間の上限 |
| `RECOVERY_MAX_TASKS_PER_PASS` | `100` | 1 回の復旧で処理する復旧タスク数の上限。最大 1000 |

`recoverRmu` と `recoverRmuWorker` は `DATABASE_URL`、`PERSISTENCE_JOURNAL_TABLE_NAME`、`PERSISTENCE_HEAD_TABLE_NAME`、`PERSISTENCE_MAX_EVENTS_PER_TRANSACTION`、`RECOVERY_MAX_TASKS_PER_PASS` を読む。コンテナイメージは同じものを使い、Lambda のイメージコマンドで `rmu` と `recoverRmuWorker` を切り替える。

## 8. 本番適用の手順

設定の正本は `deploy/rmu/event-source-mapping.json` である。IaC を持たないため、AWS CLI で適用する。`deploy/rmu/apply.sh` は既定で計画だけを表示し、`--apply` を付けたときだけ変更する。設定値の根拠と各ファイルの役割は [deploy/rmu/README.md](../deploy/rmu/README.md) にまとめている。

### Lambda の配布物を作る

RMU と復旧ワーカーは同じコンテナイメージを使い、イメージコマンドで起動するサブコマンドを切り替える。`Dockerfile.rmu` からイメージを作り、ECR に push してから両方の関数を更新する。

```sh
docker build -f Dockerfile.rmu -t <ECR のリポジトリ>:<タグ> .
docker push <ECR のリポジトリ>:<タグ>

aws lambda update-function-code --function-name <RMU の関数名> \
  --image-uri <ECR のリポジトリ>:<タグ>
aws lambda update-function-code --function-name <復旧ワーカーの関数名> \
  --image-uri <ECR のリポジトリ>:<タグ>
aws lambda wait function-updated-v2 --function-name <RMU の関数名>
aws lambda wait function-updated-v2 --function-name <復旧ワーカーの関数名>
```

`apply.sh` は、RMU のイメージコマンドを `rmu`、復旧ワーカーのイメージコマンドを `recoverRmuWorker` に設定する。ベースイメージのエントリーポイントはイメージコマンドを `_HANDLER` に渡し、バイナリはそれをサブコマンドとして使う。

### 準備

- AWS CLI v2 と `jq` を用意する。
- 両方の関数に環境変数（`DATABASE_URL`、`PERSISTENCE_*`、`RMU_RETRY_*`、`RECOVERY_MAX_TASKS_PER_PASS`、`AWS_REGION`）を設定する。`apply.sh` は環境変数を変更しない。
- 実行する IAM 主体に次の権限を許可する。
  - `lambda:ListEventSourceMappings`、`lambda:CreateEventSourceMapping`、`lambda:UpdateEventSourceMapping`、`lambda:UpdateFunctionConfiguration`、`lambda:GetFunctionConfiguration`、`lambda:AddPermission`
  - `sqs:GetQueueAttributes`、`sqs:SetQueueAttributes`
  - `events:PutRule`、`events:PutTargets`
  - `logs:CreateLogGroup`、`logs:PutMetricFilter`
  - `cloudwatch:PutMetricAlarm`

### 手順

1. 障害通知用のキューを作る。

```sh
aws sqs create-queue --queue-name cqrs-es-example-go-rmu-failure
```

2. イベントソースマッピング、キュー、スケジュールの計画を確認する。この段階では AWS を呼ばない。

```sh
./deploy/rmu/apply.sh \
  --rmu-function <RMU の関数名> \
  --head-stream-arn <head テーブルの Streams ARN> \
  --failure-queue-url <障害通知用キューの URL> \
  --recovery-function <復旧ワーカーの関数名>
```

3. `--apply` を付けて適用する。既存のイベントソースマッピングがあれば更新し、なければ作成する。EventBridge のルールも作成し、復旧ワーカーの呼び出し許可を追加する。

4. アラームとメトリクスフィルタの計画を確認する。この段階では AWS を呼ばない。

```sh
./deploy/rmu/apply-alarms.sh \
  --rmu-function <RMU の関数名> \
  --recovery-function <復旧ワーカーの関数名> \
  --failure-queue-url <障害通知用キューの URL>
```

5. `--apply` を付けてアラームを適用する。復旧ワーカーのロググループがなければ作り、完了ログから 4 つのメトリクスを作る。

```sh
./deploy/rmu/apply-alarms.sh \
  --rmu-function <RMU の関数名> \
  --recovery-function <復旧ワーカーの関数名> \
  --failure-queue-url <障害通知用キューの URL> \
  --alarm-actions <SNS トピックの ARN> \
  --apply
```

`--alarm-actions` は任意である。指定しないと、アラームはコンソールの状態だけを変え、通知を送らない。SNS トピックのポリシーで CloudWatch からの発行を許可しておく。

6. 設定を確認する。

```sh
aws lambda get-event-source-mapping --uuid <UUID> --output json
aws events list-targets-by-rule --rule cqrs-es-example-go-rmu-recovery-schedule
aws cloudwatch describe-alarms --alarm-name-prefix rmu-
```

`BisectBatchOnFunctionError` が `true`、`FunctionResponseTypes` に `ReportBatchItemFailures` が入り、`DestinationConfig.OnFailure.Destination` が障害通知用キューの ARN になっていることを確かめる。スケジュールのターゲットが復旧ワーカーの ARN を指し、アラームが `OK` または `INSUFFICIENT_DATA` で作られていることも確かめる。

7. 復旧ワーカーを 1 回手動で呼び、完了ログとメトリクスを確認する。`pending`、`deferred`、`task_failures`、`resync_failed` のフィールドが出る。

### 注意

- `update-event-source-mapping` は、ほかの更新が進行中だと失敗する。時間を置いて再実行する。
- キューを作り直すと ARN が変わる。既存のマッピングを更新するときは、同じキューを使う。
- `add-permission` は同じ statement id で再実行すると `ResourceConflictException` を返す。`apply.sh` はこの場合だけ成功として扱う。
- スケジュールの実行間隔を延ばすと、SQS の通知を失ったときに復旧タスクが残る時間が長くなる。テストは 15 分以下であることを検査する。

### ロールバック

設定を戻すときは、変更前の値で `apply.sh` と `apply-alarms.sh` をもう一度実行する。キューと `read_model_recovery_tasks` の行は消さない。未処理タスクを失うためである。

スケジュールを新しく作った場合は、次の順で外す。

```sh
aws events remove-targets --rule cqrs-es-example-go-rmu-recovery-schedule \
  --ids recovery-worker
aws events delete-rule --name cqrs-es-example-go-rmu-recovery-schedule
aws lambda remove-permission --function-name <復旧ワーカーの関数名> \
  --statement-id cqrs-es-example-go-rmu-recovery-schedule
```

アラームとメトリクスフィルタを外すときは、`apply-alarms.sh` が作った名前で削除する。

```sh
aws cloudwatch delete-alarms --alarm-names \
  rmu-recovery-pending-tasks rmu-recovery-deferred-passes \
  rmu-recovery-task-failures rmu-recovery-resync-failures \
  rmu-function-errors rmu-recovery-worker-errors rmu-failure-queue-backlog
for filter in rmu-recovery-pending-tasks rmu-recovery-deferred-passes \
  rmu-recovery-task-failures rmu-recovery-resync-failures; do
  aws logs delete-metric-filter --log-group-name /aws/lambda/<復旧ワーカーの関数名> \
    --filter-name "$filter"
done
```

イメージを戻すときは、`update-function-code` で変更前のイメージ URI を指定する。関数を削除する必要はない。

## 9. 復旧ワーカーの運用手順

### 通常の流れ

1. head Streams の処理が失敗すると、Lambda は `batchItemFailures` を返す。
2. イベントソースマッピングは失敗したレコードだけを再配送する。再試行の上限に達したレコードは OnFailure の SQS に送られる。
3. 復旧ワーカー（`recoverRmuWorker`）が SQS から通知を受ける。
4. ワーカーは 1 回の復旧を実行する。head を強整合で Scan して journal から遅れを埋め、`read_model_recovery_tasks` の `pending` を上限まで適用する。
5. 再試行できる失敗では、バッチの全メッセージを再配送対象にする。恒久的な失敗では通知を成功として扱う。復旧タスクは残るので、次の通知や定期実行で再度試せる。
6. SQS の通知がなくても、EventBridge のスケジュールが 5 分ごとに同じワーカーを起動する。上限を超えたタスクも、この定期実行が次回に処理する。

### 手動で 1 回実行する

ワーカーを空のイベントで呼ぶと、1 回の復旧だけを実行する。

```sh
aws lambda invoke --function-name <復旧ワーカーの関数名> \
  --payload '{}' --cli-binary-format raw-in-base64-out /dev/stdout
```

Lambda を使わずに同じ処理を実行するときは、`recoverRmu` コマンドを使う。

```sh
DATABASE_URL=... PERSISTENCE_JOURNAL_TABLE_NAME=journal PERSISTENCE_HEAD_TABLE_NAME=head \
  go run . recoverRmu
```

### 監視

復旧ワーカーは実行のたびに JSON の完了ログ `recovery pass finished` を出す。フィールドは `heads`、`advanced`、`skipped`、`failed`、`completed`、`pending`、`task_failures`、`deferred`、`resync_failed` である。`failed` は再同期で適用できなかった集約の数、`task_failures` は復旧タスクの再適用に失敗した数、`pending` は実行後に残った復旧タスクの数である。`deferred` と `resync_failed` は、該当するとき 1、そうでなければ 0 になる。

`apply-alarms.sh` は、このログから 4 つのメトリクスを作る。

| メトリクス | 意味 | アラーム |
|:--|:--|:--|
| `RecoveryPendingTasks` | 実行後に残った復旧タスクの数 | 15 分間 1 件以上のままなら `rmu-recovery-pending-tasks` |
| `RecoveryDeferredPasses` | 上限に達した実行か | 1 回でも 1 なら `rmu-recovery-deferred-passes` |
| `RecoveryTaskFailures` | 再適用に失敗したタスクの数 | 1 件でも出たら `rmu-recovery-task-failures` |
| `RecoveryResyncFailures` | 再同期に失敗した実行か | 1 回でも 1 なら `rmu-recovery-resync-failures` |

再試行上限の超過は、RMU の Lambda `Errors` と障害通知用キューの滞留で見る。`rmu-function-errors` は RMU の呼び出しが失敗したときに鳴る。`rmu-failure-queue-backlog` は障害通知用キューの `ApproximateNumberOfMessagesVisible` が 15 分間 1 件以上なら鳴る。復旧ワーカー自体の失敗は `rmu-recovery-worker-errors` で見る。

### 通知を失ったとき

復旧タスクは MySQL に残る。SQS のメッセージが保持期間を過ぎて消えても、EventBridge のスケジュールが 5 分以内に `pending` タスクを処理する。スケジュールを止めている環境では、`recoverRmu` を手動で実行する。SQS は通知の経路であり、正本ではない。
