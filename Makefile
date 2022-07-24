TAG?=test

build:
	@go build -o music-match main.go

test: mocks
	@go test -v -race ./...

lint:
	@golangci-lint run

docker-build:
	@docker build -t music-match:${TAG} .

mocks:
	@if ! mockgen -h 1>/dev/null 2>/dev/null; then \
  		echo "Installing mockgen..."; \
  		go install github.com/golang/mock/mockgen@v1.6.0; \
  	fi
	@echo "Generating mocks..."
	@mockgen -source neo4j/client.go -destination neo4j/client_mock.go -package neo4j