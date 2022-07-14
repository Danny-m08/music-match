FROM golang:1.18.2 AS build-env

RUN apt-get update && apt-get install -y ca-certificates openssl

ARG cert_location=/usr/local/share/ca-certificates

# Get certificate from "github.com"
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
# Get certificate from "proxy.golang.org"
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
# Update certificates
RUN update-ca-certificates


WORKDIR /usr/src/music-match
COPY go.mod go.sum ./

RUN go mod download && go mod verify
COPY . .

RUN go build -o music-match main.go


FROM centos:centos7

COPY --from=build-env /usr/src/music-match /music-match
WORKDIR /music-match

CMD ./music-match