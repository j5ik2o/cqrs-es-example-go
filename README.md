# cqrs-es-example-go

[![CI](https://github.com/j5ik2o/cqrs-es-example-go/actions/workflows/ci.yml/badge.svg)](https://github.com/j5ik2o/cqrs-es-example-go/actions/workflows/ci.yml)
[![Go project version](https://badge.fury.io/go/github.com%2Fj5ik2o%2Fcqrs-es-example-go.svg)](https://badge.fury.io/go/github.com%2Fj5ik2o%2Fcqrs-es-example-go)
[![Renovate](https://img.shields.io/badge/renovate-enabled-brightgreen.svg)](https://renovatebot.com)
[![License](https://img.shields.io/badge/License-APACHE2.0-blue.svg)](https://opensource.org/licenses/apache-2-0)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![](https://tokei.rs/b1/github/j5ik2o/cqrs-es-example-go)](https://github.com/XAMPPRocky/tokei)

## Overview

This is an example of CQRS/Event Sourcing and GraphQL implemented in Go. This example is class-based, not actor-model based.

This project uses [j5ik2o/event-store-adapter-go](https://github.com/j5ik2o/event-store-adapter-go) for Event Sourcing.

Please refer to [here](https://github.com/j5ik2o/cqrs-es-example) for implementation examples in other languages.

[日本語](./README.ja.md)

## Feature

- [x] Write API Server(GraphQL)
- [x] Read API Server(GraphQL)
- [x] Read Model Updater on Local
- [x] Docker Compose Support
- [ ] Read Model Updater on AWS Lambda
- [ ] Deployment to AWS

## Component Composition

- Write API Server
  - API is implemented by GraphQL (Mutation)
  - Event Sourced Aggregate is implemented by [j5ik2o/event-store-adapter-go](https://github.com/j5ik2o/event-store-adapter-go)
- Read Model Updater
  - Lambda to build read models based on journals
  - Locally, run code that emulates Lambda behavior (local-rmu)
- Read API Server
  - API is implemented by GraphQL (Query)

## Stack

This OSS repository mainly utilizes the following technology stack.

- [99designs/gqlgen](https://github.com/99designs/gqlgen)
- [jmoiron/sqlx](https://github.com/jmoiron/sqlx)
- [j5ik2o/event-store-adapter-go](https://github.com/j5ik2o/event-store-adapter-go)

## System Architecture Diagram

![](docs/images/system-layout.png)

## Development Environment

- [Tools Setup](docs/TOOLS_INSTALLATION.md)
- [Build and Test](docs/BUILD_AND_TEST.md)

### Local Environment

- [Debugging with Docker Compose](docs/DEBUG_ON_DOCKER_COMPOSE.md)

## Links

- [Common Documents](https://github.com/j5ik2o/cqrs-es-example)
