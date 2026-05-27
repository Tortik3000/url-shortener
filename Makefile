MODULE=github.com/edkin/url-shortener

PROJECT_ROOT := $(CURDIR)
LOCAL_BIN := $(PROJECT_ROOT)/bin

.PHONY: generate
generate:
	sqlc generate
	go generate ./...
	go mod tidy

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: lint
lint:
	go vet ./...

COVER_PKGS := $(shell go list ./... | grep -Ev '(/sqlc$$|/mocks$$|/cmd/)' | tr '\n' ',' | sed 's/,$$//')

.PHONY: test
test:
	go test -race -coverprofile=coverage.out -coverpkg=$(COVER_PKGS) ./...
	@go tool cover -func=coverage.out | grep "^total:"

.PHONY: build
build:
	go build -o $(LOCAL_BIN)/server ./cmd/server

.PHONY: up
up:
	docker compose up --build

.PHONY: down
down:
	docker compose down
