SERVICE_NAME=go-api-demo
MYSQL_DSN=mysql://user:password@tcp(127.0.0.1:3306)/go-api-demo
MYSQL_MIGRATION_PATH=internal/storage/mysql/migrations
export HOST_IP=${shell ipconfig getifaddr en0}

################################################################################

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

.PHONY: test-unit
test-unit: lint
	go test -v -race -bench=./... -benchmem -timeout=120s -cover -coverprofile=./test/coverage.txt `go list ./... | grep -v test`

.PHONY: test-e2e
test-e2e: lint
	go test -v -race -bench=./... -benchmem -timeout=120s -cover -coverprofile=./test/coverage.txt `go list ./... | grep test`

.PHONY: .env
.env:
ifeq (,$(wildcard .env))
	cp .env.dist .env
endif

################################################################################

.PHONY: docker-up
docker-up: .env
	docker compose -f docker/dev/docker-compose.yml up -d

.PHONY: docker-down
docker-down:
	docker rm --force -v go-api-demo-connect go-api-demo-grafana go-api-demo-prometheus go-api-demo-jaeger go-api-demo-redis go-api-demo-elastic go-api-demo-zookeeper go-api-demo-kowl go-api-demo-db go-api-demo-kafka go-api-demo-schema-registry
	docker compose -f docker/dev/docker-compose.yml down

################################################################################

.PHONY: migrate-up
migrate-up:
	migrate -database '${MYSQL_DSN}' -path '${MYSQL_MIGRATION_PATH}' -verbose up

.PHONY: migrate-down
migrate-down:
	migrate -database '${MYSQL_DSN}' -path '${MYSQL_MIGRATION_PATH}' -verbose down

.PHONY: migrate-drop
migrate-drop:
	migrate -database '${MYSQL_DSN}' -path '${MYSQL_MIGRATION_PATH}' -verbose drop
