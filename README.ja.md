# cqrs-es-example-go

## 概要

これはGoで実装されたCQRS/Event Sourcingの例です。

このプロジェクトは、イベントソーシングのために[j5ik2o/event-store-adapter-go](j5ik2o/event-store-adapter-go)を使用しています。

[English](./README.md)

## フィーチャー

- [x] Write API Server(REST)の実装
- [x] Read API Server(GraphQL)の実装
- [x] Local版のRead Model Updaterの実装
- [x] Docker Composeによる動作確認
- [ ] AWS Lambda版のRead Model Updaterの実装
- [ ] AWSへのデプロイ

## コンポーネント構成

- Write APIサーバ: 書き込み専用Web API
- Read Model Updater: ジャーナルに基づいて読み取りモデルを構築する非同期プロセス
- Read API Server: GraphQLによって実装された読み取り専用API

## システムアーキテクチャ図

![](docs/images/system-layout.png)

## 使い方

### ローカル環境

#### サービスの起動　

```shell
$ make docker-compose-up
./tools/scripts/docker-compose-up.sh
ARCH=arm64
 Container read-api-server-1  Stopping
...
 Container docker-compose-migration-1  Started
 Container read-api-server-1  Starting
 Container read-api-server-1  Started
```

#### APIのテスト

```shell
$ make verify-group-chat
ADMIN_ID="UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z" \
        WRITE_API_SERVER_BASE_URL=http://localhost:18080 \
        READ_API_SERVER_BASE_URL=http://localhost:18082 \
        ./tools/scripts/verify-group-chat.sh
{"group_chat_id":"GroupChat-01HPG4EV94HMPT08GZS0ZWW0VJ"}
GroupChat:
{
  "data": {
    "getGroupChat": {
      "id": "GroupChat-01HPG4EV94HMPT08GZS0ZWW0VJ",
      "name": "group-chat-example-1",
      "ownerId": "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z",
      "createdAt": "2024-02-13 02:24:12 +0000 UTC",
      "updatedAt": "2024-02-13 02:24:12 +0000 UTC"
    }
  }
}
...
```

## 参考リンク

- [for Rust](https://github.com/j5ik2o/cqrs-es-example-rs)
- [Common Documents](https://github.com/j5ik2o/cqrs-es-example-docs)
