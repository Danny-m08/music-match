TAG?=test

build:
	@go build -o music-match main.go

test:
	@go test -v -race ./...

docker-build:
	@docker build -t music-match:${TAG} .