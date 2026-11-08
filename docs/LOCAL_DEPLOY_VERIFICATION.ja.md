# デプロイ資産のローカル検証

この文書は、`deploy/rmu` のイベントソースマッピング・アラーム定義・適用スクリプトを、実 AWS リソースを作らずにローカルでどこまで検証するかを定める。対象は issue #29 の完了条件のうち、ローカルのコンテナ検証までの範囲である。

実 AWS へのデプロイ、Lambda 関数の作成、Secrets Manager / SNS の実設定、公開・リリースは行わない。

## 1. 検証の構成

`make docker-compose-ci` を実行すると、Docker Compose と LocalStack の上で次の順に検証する。

1. Write API → RMU → Read API の一連の流れ。
2. RMU を再起動したときの replay と、行が重複しないこと。
3. 復旧ワーカーの起動パス（SQS OnFailure 相当）と定期実行パス（EventBridge スケジュール相当）。
4. `deploy/rmu` の JSON・シェル構文、適用スクリプトの dry-run、LocalStack に対する AWS CLI の適用。

CI の `e2e` ジョブは `make docker-compose-ci` を実行する。ローカルでも同じコマンドで再現できる。

## 2. Write API → RMU → Read API

`tools/e2e-test/verify-group-chat.sh` が GraphQL のミューテーションとクエリを順に実行する。Read API は RMU の投影が追いつくまでポーリングするため、固定の待ち時間に依存しない。

再起動の検証は `tools/scripts/docker-compose-e2e-test.sh` が行う。

1. 再起動前の `read_model_positions` とリードモデルの行数を控える。
2. `read-model-updater-1` を再起動する。RMU は head Streams を `TRIM_HORIZON` から読み直す。
3. 再起動後に新しいグループチャットとメッセージを書き、Read API に反映されるまで待つ。
4. 再起動前の位置が変わっておらず、行数が増えたのが新しい書き込みの分だけであることを確かめる。

再送された head レコードは `read_model_positions` の位置ガードで重複として捨てられる。これが「再起動後の replay と重複なし」の確認である。

## 3. 復旧ワーカー

### ローカルワーカーの動き

`localRecoveryWorker` コマンドは、Lambda の復旧ワーカーと同じ `rmu.RecoveryWorker` を呼ぶ。

- 起動時に 1 回、`HandleSQSEvent` を呼ぶ。これは SQS OnFailure の通知を受けたときと同じ経路で、head を強整合で Scan して journal から遅れを埋め、`read_model_recovery_tasks` の `pending` を再適用する。
- その後、`RECOVERY_WORKER_INTERVAL`（既定 5 秒、compose では 2 秒）ごとに `HandleScheduledEvent` を呼ぶ。これは EventBridge の定期実行と同じ経路で、`pending` タスクだけを再適用する。
- 完了ログは Lambda 版と同じ JSON の `recovery pass finished` である。`pending`、`deferred`、`task_failures`、`resync_failed` をメトリクスフィルタと同じ形で出す。

Lambda を使わないため、SQS のメッセージ配送そのものは再現しない。メッセージ本文を読まず、MySQL の復旧タスクを正本にするという契約は同じコードで検証する。

### `tools/scripts/docker-compose-recovery-test.sh` の検証

1. 再起動検証で作ったグループチャットを対象にする。
2. RMU と復旧ワーカーを止め、リードモデルの行と `read_model_positions` を消して、投影が失われた状態を作る。
3. 恒久的な失敗が残すのと同じ形で、`read_model_recovery_tasks` に `pending` のタスクを 1 件書く。`store_id` は head テーブルの `__config__` から読む。
4. 復旧ワーカーを起動する。起動パスが journal からリードモデルを復元し、タスクを `completed` にする。
5. 行数と位置が元に戻り、重複した行がないこと、`pending` が 0 であることを確かめる。
6. 完了ログが増え続けることで、定期実行が動いていることを確かめる。直近の完了ログの `pending` が 0、`resync_failed` が 0 であることも見る。
7. `recoverRmu` を手動でもう 1 回実行し、二重適用がないことを確かめる。
8. RMU を起動して元の状態に戻す。

## 4. `deploy/rmu` の検証

`tools/e2e-test/verify-deploy-assets.sh` が e2e-test コンテナの中で実行する。

### 構文と設定値

- `event-source-mapping.json` と `cloudwatch-alarms.json` を `jq empty` で構文検査する。
- `apply.sh` と `apply-alarms.sh` を `bash -n` で構文検査する。
- Go のテスト `deploy/rmu/event_source_mapping_test.go` と `cloudwatch_alarms_test.go` が、未知のフィールドを拒否しつつ、部分バッチ応答、タイムアウトと再試行の関係、Lambda / SQS の上限、定期実行の間隔、アラームの網羅を検査する。
- `deploy/rmu/apply_scripts_test.go` が、偽の `aws` を PATH に置いて dry-run を実行し、dry-run が AWS を呼ばないことと、表示される計画値が JSON と一致することを検査する。

### dry-run

`apply.sh` と `apply-alarms.sh` は `--apply` を付けない限り AWS を呼ばない。検証スクリプトは `AWS_ENDPOINT_URL` を閉じたポートに向けて実行し、呼び出しがあれば失敗するようにしている。

- イベントソースマッピング: `batch 100`、`bisect true`、`ReportBatchItemFailures`、再試行 3 回、最大経過 86400 秒。
- 障害通知キュー: 保持 1209600 秒、可視性タイムアウト 960 秒。
- 復旧ワーカー: イメージコマンド `recoverRmuWorker`、タイムアウト 900 秒。
- スケジュール: `rate(5 minutes)`、`ENABLED`。

### LocalStack への接続

LocalStack の SQS / EventBridge / CloudWatch Logs / CloudWatch に対して、設定値と同じコマンドを実行する。

- 障害通知キューを作り、`MessageRetentionPeriod` と `VisibilityTimeout` を設定して読み戻す。
- スケジュールのルールを設定値で作り、`ScheduleExpression` と `State` を読み戻す。
- `apply-alarms.sh --apply` を実行し、4 つのメトリクスフィルタと 7 つのアラームが作られることを確かめる。アラームアクションは指定しない。SNS トピックを作らないためである。
- `rmu-recovery-pending-tasks` の `EvaluationPeriods` が JSON と一致することを確かめる。

### ローカルで検証できないこと

次は実 AWS の Lambda とアカウントが必要なので、この検証の対象外である。設定値そのものは Go のテストと dry-run で検査し、本番適用時に `aws lambda get-event-source-mapping` などで確認する。

- Lambda のイベントソースマッピングが `FunctionResponseTypes` を解釈し、返した `batchItemFailures` だけを再配送すること。
- `BisectBatchOnFunctionError` と `DestinationConfig.OnFailure` による実際の SQS 配送。
- Lambda のイメージコマンド切り替えと `_HANDLER` の解決。
- SNS によるアラーム通知と Secrets Manager の設定。

## 5. 個別に実行する

スタックを起動したまま個別に確認できる。

```sh
# Write API → RMU → Read API と再起動
make docker-compose-ci

# 復旧ワーカーの検証だけを再実行
./tools/scripts/docker-compose-recovery-test.sh

# deploy/rmu の検証だけを再実行
docker compose -p cqrs-es-example-go \
  -f tools/docker-compose/docker-compose-databases.yml \
  -f tools/docker-compose/docker-compose-applications.yml \
  -f tools/docker-compose/docker-compose-e2e-test.yml \
  run -T --rm --no-deps --entrypoint ./verify-deploy-assets.sh e2e-test
```

復旧の状態は MySQL で確認できる。

```sh
docker compose -p cqrs-es-example-go \
  -f tools/docker-compose/docker-compose-databases.yml \
  exec -T mysql mysql --protocol=TCP -h127.0.0.1 -uceer -pceer ceer \
  -e "SELECT aggregate_id, seq_nr FROM read_model_positions; \
      SELECT status, COUNT(*) FROM read_model_recovery_tasks GROUP BY status;"
```

復旧ワーカーの完了ログは次のコマンドで見る。

```sh
docker compose -p cqrs-es-example-go \
  -f tools/docker-compose/docker-compose-databases.yml \
  -f tools/docker-compose/docker-compose-applications.yml \
  logs --no-color recovery-worker | grep 'recovery pass finished'
```

## 6. 関連ファイル

- `cmd/localRecoveryWorker.go`: Lambda 外で動く復旧ワーカー。
- `tools/scripts/docker-compose-recovery-test.sh`: 復旧経路の e2e。
- `tools/e2e-test/verify-deploy-assets.sh`: デプロイ資産の検証。
- `deploy/rmu/apply_scripts_test.go`: 適用スクリプトの dry-run と構文のテスト。
- `deploy/rmu/event-source-mapping.json`: イベントソースマッピングの正本。
- `deploy/rmu/cloudwatch-alarms.json`: アラームとメトリクスフィルタの正本。
- `docs/RMU_RECOVERY.ja.md`: 復旧の設計と本番適用の手順。
