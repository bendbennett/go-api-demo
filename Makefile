SERVICE_NAME=go-api-demo

.PHONY: fmt
fmt:
	gofmt -w -s .
	goimports -w .
	go clean ./...

GREEN=\033[0;32m # Green
NC=\033[0m # No Color

.PHONY: lint
lint:
	golangci-lint run --config golangci.yml --timeout=10m
	@echo "${GREEN}[✔] golangci-lint OK${NC}"

.PHONY: proto
proto:
	protoc \
		-I=proto \
		--go_out=generated \
		--go_opt=paths=source_relative \
		--go-grpc_out=generated \
		--go-grpc_opt=paths=source_relative \
		user.proto

.PHONY: build
build:
	go build \
		-ldflags "-w -s -X github.com/bendbennett/go-api-demo/internal/app.commitHash=`git rev-parse HEAD`" \
		-race \
		-o ./bin/$(SERVICE_NAME) \
		-v ./cmd/main.go

.PHONY: run
run: build
	bin/$(SERVICE_NAME)

.PHONY: test
test: lint
	go test -v -race -bench=./... -benchmem -timeout=120s -cover -coverprofile=./test/coverage.txt ./...
