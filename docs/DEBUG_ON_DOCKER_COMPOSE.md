# Debugging on Docker Compose

## Build the image

```shell
$ make docker-build
```

## Start docker-compose

```shell
$ make docker-compose-up
```

The required database and tables will be created and the application will be started.
If you want to start only the database for development purposes, run `docker-compose-up-db` instead of `docker-compose-up`.

## Stop docker-compose

```shell
$ make docker-compose-down
```

## Verification

Run the verification script against the running services from the host. It waits for the Write API and Read API to answer, then runs the GroupChat flow from creation to deletion.

```shell
$ make verify-group-chat
```

## E2E tests

Build the images, start the containers, and run the end-to-end test in one command. The CI e2e job runs the same command.

```shell
$ make docker-compose-ci
```

`docker-compose-ci` proceeds in this order and returns a non-zero exit code at the first failure.

1. Build the application and e2e test images.
2. Start LocalStack, MySQL, migration, dynamodb-setup, the Write API, the RMU, the Read API, and the e2e-test container.
3. Wait for the exit code of the e2e-test container. The verification script polls until the Write API and Read API answer, and polls the read model instead of sleeping for a fixed time.
4. Restart the RMU so it replays the head stream from the beginning. Wait until a write made after the restart reaches the read model, then check that the `read_model_positions` rows and the read model row counts did not change. This confirms the replayed events are dropped as duplicates.

The Write API listens on 28080, the RMU on 28081, and the Read API on 28082. The e2e test sends GraphQL requests to `/query`.

