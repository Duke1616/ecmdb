FROM golang:1.21.1-alpine as build
MAINTAINER luankz
WORKDIR /app
COPY . .
ENV GOPROXY="https://goproxy.cn,direct"
# ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go mod tidy
RUN go build -a -o dist/ecmdb -ldflags="-w -s" main.go
FROM alpine:latest as deploy
ENV EGO_DEBUG=true
WORKDIR /app
COPY --from=build /app/dist/ecmdb .
EXPOSE 8000