FROM golang:1.18.2 AS build-env

ENV GO111MODULE=on

RUN mkdir /music-match
WORKDIR /music-match
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build


FROM centos:centos7

COPY --from=build-env /music-match /music-match
WORKDIR /music-match

CMD ./music-match