FROM golang:1.23.0-alpine as build
MAINTAINER luankz
WORKDIR /app
COPY . .
# ENV GOPROXY="https://goproxy.cn,direct"
# ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go mod tidy
ARG VERSION=unknown
RUN go build -ldflags="-X main.version=${VERSION} -w -s" -a -o dist/ecmdb main.go
FROM alpine:latest as deploy
ENV EGO_DEBUG=true
WORKDIR /app
COPY --from=build /app/dist/ecmdb .
EXPOSE 8000