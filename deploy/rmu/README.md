# RMU deployment assets

This directory holds the production settings for the Read Model Updater (RMU)
Lambda function and its SQS recovery worker. The repository has no Terraform,
CDK, or SAM stack, so the settings live in `event-source-mapping.json` and are
applied with the AWS CLI through `apply.sh`. The Japanese procedure and the
reasoning behind each value are in [RMU recovery](../../docs/RMU_RECOVERY.ja.md).

## Files

- `event-source-mapping.json` is the source of truth for the head stream
  mapping, the failure queue, and the recovery worker mapping. A Go test in
  this directory validates every safety setting, so a change that disables
  partial batch responses or shrinks the timeout fails CI.
- `apply.sh` reads that file and applies it. It is a dry run unless `--apply`
  is passed.

## Prerequisites

- AWS CLI v2 and `jq`.
- Two Lambda functions built from the container image in `Dockerfile.rmu`.
  The main function uses the image command `rmu`; the recovery worker uses
  `recoverRmuWorker`. The base image entrypoint passes the image command
  through `_HANDLER`, and the binary uses it as the subcommand.
- The functions already have their environment variables set. The script
  changes the timeout, the image command, the queue attributes, and the event
  source mappings. It does not manage environment variables, IAM roles, VPC
  settings, or the database connection.

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
none exists and updates the existing one otherwise.

## Verify

```sh
aws lambda get-function-configuration --function-name cqrs-es-example-go-rmu \
  --query '{Timeout:Timeout,ImageConfig:ImageConfig}' --output json
aws lambda get-event-source-mapping --uuid <uuid> --output json
aws sqs get-queue-attributes --queue-url <queue-url> \
  --attribute-names All --output json
```

The mapping must report `BisectBatchOnFunctionError: true`,
`FunctionResponseTypes: ["ReportBatchItemFailures"]`, and an `OnFailure`
destination that points at the failure queue.

## Recovery worker

The stream mapping sends a failed batch to the failure queue. The worker
consumes the queue and runs one recovery pass: it reconciles every head from
the journal, then replays the pending tasks in `read_model_recovery_tasks`.
The message body is ignored, so a redelivered or batched message is harmless.
A retryable failure returns every message ID as a batch item failure; a
permanent failure is acknowledged because the task is already stored.

The worker uses the image command `recoverRmuWorker` and the same
`DATABASE_URL`, `PERSISTENCE_*`, and `AWS_REGION` variables as the main
function. To run one pass by hand, invoke the function with an empty SQS
event (`{}`). The one-shot CLI `recoverRmu` runs the same code outside Lambda.
