# Docker Compose上でデバッグする

## ビルドイメージ

```shell
$ make docker-build
```

## docker-composeの起動

```shell
$ make docker-compose-up
```

必要なデータベースとテーブルが作成され、アプリケーションも起動します。
開発目的でデータベースだけを起動したい場合は、`docker-compose-up`ではなく`docker-compose-up-db`を実行してください。

## docker-composeの停止

```shell
$ make docker-compose-down
```

## 動作確認

起動中のサービスに対して、ホストから検証スクリプトを実行します。Write API と Read API の応答を待ってから、GroupChat の作成から削除までを順に実行します。

```shell
$ make verify-group-chat
```

## e2e テスト

イメージのビルドからコンテナの起動、e2e の実行までをまとめて行います。CI の e2e ジョブもこのコマンドを実行します。

```shell
$ make docker-compose-ci
```

`docker-compose-ci` は次の順に進み、途中で失敗した時点で非ゼロの終了コードを返します。

1. アプリケーションと e2e テストのイメージをビルドする。
2. LocalStack、MySQL、migration、dynamodb-setup、Write API、RMU、Read API、e2e-test を起動する。
3. e2e-test コンテナの終了コードを待つ。検証スクリプトは Write API と Read API が応答するまでポーリングし、Read Model への反映もポーリングで待つ。
4. RMU を再起動し、head Streams を最初から読み直させる。再起動後の書き込みが Read Model に反映されるまで待ち、再起動前の `read_model_positions` とリードモデルの行数が変わっていないことを確認する。これで、再送されたイベントが重複として捨てられることを確かめる。

Write API は 28080、RMU は 28081、Read API は 28082 で待ち受けます。e2e の検証は GraphQL の `/query` に対して行います。

