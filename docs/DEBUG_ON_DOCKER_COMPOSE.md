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

```shell
$ make verify-group-chat
```
