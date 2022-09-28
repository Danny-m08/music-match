.EXPORT_ALL_VARIABLES:

TAG?=test
CONFIG_PATH?=config-local.yaml

build:
	@go build -o music-match main.go

test:
	@go test -v -race ./...

lint:
	@golangci-lint run

docker-build:
	@docker build -t music-match:${TAG} .

deploy: build
	@if ! docker ps | grep neo4j 2>&1 1>/dev/null; then \
		docker compose up -d; \
	fi
	@./music-match