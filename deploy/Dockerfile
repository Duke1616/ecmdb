FROM golang:1.21.1-alpine as build
MAINTAINER luankz
WORKDIR /usr/src/backend
COPY . .
ENV GOPROXY="https://goproxy.cn,direct"
ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go mod tidy
RUN go build -a -o dist/ecmdb -ldflags="-w -s" main.go wire_gen.go
FROM alpine:latest as deploy
WORKDIR /usr/src/backend
COPY --from=build /usr/src/backend/dist/ecmdb .
EXPOSE 8000