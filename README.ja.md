# cqrs-es-example-go

## 概要

これはGoで実装されたCQRS/Event Sourcing + GraphQLの例です。

このプロジェクトは、イベントソーシングのために[j5ik2o/event-store-adapter-go](j5ik2o/event-store-adapter-go)を使用しています。

[English](./README.md)

## フィーチャー

- [x] Write API Server(GraphQL)の実装
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

## 開発環境

- [ツールのセットアップ](docs/TOOLS_INSTALLATION.md)

### ローカル環境

- [Docker Composeでのデバッグ](docs/DEBUG_ON_DOCKER_COMPOSE.md)

## 参考リンク

- [Rust版](https://github.com/j5ik2o/cqrs-es-example-rs)
- [TypeScript版](https://github.com/j5ik2o/cqrs-es-example-js)
- [共通ドキュメント](https://github.com/j5ik2o/cqrs-es-example-docs)
