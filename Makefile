TAG?=test

build:
	@go build -o music-match main.go

test:
	@go test -v -race ./...

lint:
	@golangci-lint run

docker-build:
	@docker build -t music-match:${TAG} .

mocks:
	 mockgen -source neo4j/client.go -destination neo4j/client_mock.go -package neo4j