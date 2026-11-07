# RMU deployment assets

This directory holds the production settings for the Read Model Updater (RMU)
Lambda function and its SQS recovery worker. The repository has no Terraform,
CDK, or SAM stack, so the settings live in `event-source-mapping.json` and are
applied with the AWS CLI through `apply.sh`. The Japanese procedure and the
reasoning behind each value are in [RMU recovery](../../docs/RMU_RECOVERY.ja.md).

## Files

- `event-source-mapping.json` is the source of truth for the head stream
  mapping, the failure queue, the recovery worker mapping, and the EventBridge
  schedule that acts as the safety net when an SQS notification is lost. A Go
  test in this directory validates every safety setting, so a change that
  disables partial batch responses, shrinks the timeout, or slows the schedule
  fails CI.
- `apply.sh` reads that file and applies it. It is a dry run unless `--apply`
  is passed.
- `cloudwatch-alarms.json` defines the metric filters and alarms for a pending
  task backlog, a deferred pass, recovery failures, resync failures, retry
  exhaustion, and a failure queue backlog. Another Go test validates it.
- `apply-alarms.sh` reads that file and applies it. It is a dry run unless
  `--apply` is passed.

## Prerequisites

- AWS CLI v2 and `jq`.
- Two Lambda functions built from the container image in `Dockerfile.rmu`.
  The main function uses the image command `rmu`; the recovery worker uses
  `recoverRmuWorker`. The base image entrypoint passes the image command
  through `_HANDLER`, and the binary uses it as the subcommand. Build, push,
  and update both functions with `aws lambda update-function-code`.
- The functions already have their environment variables set, including
  `RECOVERY_MAX_TASKS_PER_PASS` (default `100`, maximum `1000`). The script
  changes the timeout, the image command, the queue attributes, the event
  source mappings, and the schedule. It does not manage environment variables,
  IAM roles, VPC settings, or the database connection.

## Usage

Create the failure queue once. The queue receives failed head stream batches
and wakes the recovery worker.

```sh
aws sqs create-queue --queue-name cqrs-es-example-go-rmu-failure
```

Plan the change. This makes no AWS calls.

```sh
./deploy/rmu/apply.sh \
  --rmu-function cqrs-es-example-go-rmu \
  --head-stream-arn arn:aws:dynamodb:ap-northeast-1:123456789012:table/head/stream/2026-01-01T00:00:00.000 \
  --failure-queue-url https://sqs.ap-northeast-1.amazonaws.com/123456789012/cqrs-es-example-go-rmu-failure \
  --recovery-function cqrs-es-example-go-rmu-recovery
```

Apply it by adding `--apply`. The script creates an event source mapping when
none exists and updates the existing one otherwise. It also creates the
EventBridge rule, points it at the recovery worker, and grants
`events.amazonaws.com` permission to invoke the function.

Apply the alarms in the same way.

```sh
./deploy/rmu/apply-alarms.sh \
  --rmu-function cqrs-es-example-go-rmu \
  --recovery-function cqrs-es-example-go-rmu-recovery \
  --failure-queue-url https://sqs.ap-northeast-1.amazonaws.com/123456789012/cqrs-es-example-go-rmu-failure \
  --alarm-actions arn:aws:sns:ap-northeast-1:123456789012:cqrs-es-example-go-rmu-alerts
```

Add `--apply` to create the recovery worker log group, the metric filters, and
the alarms. The metric filters read the JSON field values from the
`recovery pass finished` log line. `--alarm-actions` is optional; without it
the alarms change state in the console but send no notification.

## Verify

```sh
aws lambda get-function-configuration --function-name cqrs-es-example-go-rmu \
  --query '{Timeout:Timeout,ImageConfig:ImageConfig}' --output json
aws lambda get-event-source-mapping --uuid <uuid> --output json
aws sqs get-queue-attributes --queue-url <queue-url> \
  --attribute-names All --output json
aws events list-targets-by-rule --rule cqrs-es-example-go-rmu-recovery-schedule
aws cloudwatch describe-alarms --alarm-name-prefix rmu-
```

The mapping must report `BisectBatchOnFunctionError: true`,
`FunctionResponseTypes: ["ReportBatchItemFailures"]`, and an `OnFailure`
destination that points at the failure queue. The schedule target must be the
recovery worker's ARN. Invoke the worker with an empty payload to produce one
completion log line.

## Recovery worker

The stream mapping sends a failed batch to the failure queue. The worker
consumes the queue and runs one recovery pass: it reconciles every head from
the journal, then replays the pending tasks in `read_model_recovery_tasks`.
The message body is ignored, so a redelivered or batched message is harmless.
A retryable failure returns every message ID as a batch item failure; a
permanent failure is acknowledged because the task is already stored.

The worker uses the image command `recoverRmuWorker` and the same
`DATABASE_URL`, `PERSISTENCE_*`, `RECOVERY_MAX_TASKS_PER_PASS`, and `AWS_REGION`
variables as the main function. To run one pass by hand, invoke the function
with an empty payload (`{}`). The one-shot CLI `recoverRmu` runs the same code
outside Lambda.

One pass processes at most `RECOVERY_MAX_TASKS_PER_PASS` tasks (default `100`,
maximum `1000`). When more tasks are pending, the rest stay `pending` and the
next run picks them up. The completion log records `deferred: 1` and the
remaining `pending` count. The EventBridge rule runs the worker every five
minutes even when no SQS notification arrives.

## Rollback

Re-run `apply.sh` and `apply-alarms.sh` with the previous values. Do not delete
the failure queue or the rows in `read_model_recovery_tasks`: either would lose
work that has not been applied yet.

If the schedule was added by this deployment, remove the target, delete the
rule, and remove the Lambda permission. If the alarms were added, delete them
and their metric filters by name. Restore the previous image with
`aws lambda update-function-code`; deleting the functions is not required. The
Japanese procedure in [RMU recovery](../../docs/RMU_RECOVERY.ja.md) carries the
exact commands.
