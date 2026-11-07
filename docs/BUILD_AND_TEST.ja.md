## ビルド方法

ビルドを行う前に、Dockerを必ず起動してください。

```shell
$ make prepare # 初回のみ 
$ make build
```

## テスト方法

テストを実行する前に、Dockerを必ず起動してください。

```shell
$ make test
```

## Lint方法

`make vet` は `go vet` を、`make lint` は [staticcheck](https://staticcheck.dev/) を実行する。CI の `lint` ジョブでは `make vet` と `make lint` の両方を実行する。

```shell
$ make vet
$ make lint
```
