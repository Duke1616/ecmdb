FROM golang:1.25.0-alpine as build
MAINTAINER luankz
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY . .
RUN go mod download
ARG VERSION=unknown
RUN go build -ldflags="-X main.version=${VERSION} -w -s" -a -o dist/ecmdb main.go
FROM alpine:latest as deploy
ENV EGO_DEBUG=true
WORKDIR /app
COPY --from=build /app/dist/ecmdb .
ENTRYPOINT ["./ecmdb", "server"]