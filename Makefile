# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run

WRITE_API_SERVER_BASE_URL=http://localhost:28080
READ_API_SERVER_BASE_URL=http://localhost:28082

.PHONY: all
all: test

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: clean
clean:
	$(GOMOD) tidy
	rm -rf ./testdata

.PHONY: deps
deps:
	$(GOGET) -u

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: vet
vet:
	$(GOCMD) vet ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Run all code checks
.PHONY: check
check: lint vet fmt test

# Download all dependencies
prepare:
	$(GOMOD) download

# Update all dependencies
update:
	$(GOMOD) tidy
	$(GOMOD) download

build:
	$(GOBUILD)

c-gql-init:
	@echo "Initializing GraphQL code for command..."
	$(GORUN) github.com/99designs/gqlgen init --config pkg/command/interfaceAdaptor/gqlgen.yml
	$(GOMOD) tidy

c-gql-gen:
	@echo "Generating GraphQL code for command..."
	$(GORUN) github.com/99designs/gqlgen generate --config pkg/command/interfaceAdaptor/gqlgen.yml
	$(GOMOD) tidy

q-gql-init:
	@echo "Initializing GraphQL code..."
	$(GORUN) github.com/99designs/gqlgen init --config pkg/query/interfaceAdaptor/gqlgen.yml
	$(GOMOD) tidy

q-gql-gen:
	@echo "Generating GraphQL code..."
	$(GORUN) github.com/99designs/gqlgen generate --config pkg/query/interfaceAdaptor/gqlgen.yml
	$(GOMOD) tidy

.PHONY: run-write-api-server
run-write-api-server:
	AWS_REGION=ap-northeast-1 \
	AWS_DYNAMODB_ENDPOINT_URL=http://localhost:28000 \
	AWS_DYNAMODB_ACCESS_KEY_ID=x \
	AWS_DYNAMODB_SECRET_ACCESS_KEY=x \
		$(GORUN) main.go writeApi

.PHONY: run-read-api-server
run-read-api-server:
	AWS_REGION=ap-northeast-1 \
	DATABASE_URL='ceer:ceer@tcp(localhost:23306)/ceer' \
		$(GORUN) main.go readApi

.PHONY: docker-build
docker-build:
	docker build -t cqrs-es-example-go:latest -f Dockerfile .

.PHONY: docker-build-rmu
docker-build-rmu:
	docker build -t cqrs-es-example-go-rmu:latest -f Dockerfile.rmu .

.PHONY: docker-compose-up
docker-compose-up:
	./tools/scripts/docker-compose-up.sh

.PHONY: docker-compose-up-db
docker-compose-up-db:
	./tools/scripts/docker-compose-up.sh -d

.PHONY: docker-compose-ps
docker-compose-ps:
	./tools/scripts/docker-compose-ps.sh

.PHONY: docker-compose-down
docker-compose-down:
	./tools/scripts/docker-compose-down.sh

.PHONY: verify-group-chat
verify-group-chat:
	./tools/e2e-test/verify-group-chat.sh

.PHONY: docker-build-e2e-test
docker-build-e2e-test:
	cd ./tools/e2e-test && make build

.PHONY: docker-compose-e2e-test
docker-compose-ci: docker-build-e2e-test docker-build
	./tools/scripts/docker-compose-e2e-test.sh
