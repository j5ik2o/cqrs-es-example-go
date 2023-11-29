# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build

WRITE_API_SERVER_BASE_URL=http://localhost:18080
READ_API_SERVER_BASE_URL=http://localhost:18082

.PHONY: install-tools
install-tools:
	@which staticcheck > /dev/null || go install honnef.co/go/tools/cmd/staticcheck@latest
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest

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

q-gql-init:
	@echo "Generating GraphQL code..."
	@go run github.com/99designs/gqlgen init --config pkg/query/gqlgen.yml
	$(GOMOD) tidy

q-gql-gen:
	@echo "Generating GraphQL code..."
	@go run github.com/99designs/gqlgen generate --config pkg/query/gqlgen.yml
	$(GOMOD) tidy

.PHONY: docker-build
docker-build:
	docker build -t cqrs-es-example-go:latest -f Dockerfile .

.PHONY: docker-build-rmu
docker-build-rmu:
	docker build -t cqrs-es-example-go-rmu:latest -f Dockerfile.rmu .

.PHONY: docker-compose-up
docker-compose-up:
	./tools/scripts/docker-compose-up.sh

.PHONY: docker-compose-down
docker-compose-down:
	./tools/scripts/docker-compose-down.sh

.PHONY: create-group-chat
create-group-chat:
	WRITE_API_SERVER_BASE_URL=$(WRITE_API_SERVER_BASE_URL) ./tools/scripts/curl-post-create-group-chat.sh

.PHONY: get-group-chat
get-group-chat:
	READ_API_SERVER_BASE_URL=$(READ_API_SERVER_BASE_URL) ./tools/scripts/curl-post-get-group-chat.sh
