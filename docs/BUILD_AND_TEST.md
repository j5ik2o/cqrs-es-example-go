## How to build

Be sure to start Docker before performing a build.

```shell
$ make prepare # first time only 
$ make build
```

## How to test

Be sure to start Docker before running the test.

```shell
$ make test
```

## How to lint

`make vet` runs `go vet`, and `make lint` runs [staticcheck](https://staticcheck.dev/).
The CI `lint` job runs both `make vet` and `make lint`.

```shell
$ make vet
$ make lint
```
