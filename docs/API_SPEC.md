# API仕様の確認方法

## Write API Server: OpenAPIの仕様を確認する方法

### 定義ファイルで確認する方法

`docs/swagger.json`もしくは`docs/swagger.yaml`を確認してください。

コード上のアノテーションを変更をした場合は以下のコマンドで定義ファイルを更新してください。

```shell
$ make swag
```

### Swagger UIで確認する場合

まずはWrite API Serverを起動する

```shell
$ make docker-compose-up-db
$ make run-write-api-server
```

もしくは

```shell
$ make docker-compose-build
$ make docker-compose-up
```

以下のコマンドでswagger-uiをブラウザで開きます。

```shell
$ make view-swagger-ui
```

## Read API Server: GraphQLのSDLを確認する方法

`pkg/query/graph/schema.graphqls`を確認してください。

