FROM golang:1.23.0-alpine as build
MAINTAINER luankz
WORKDIR /app
COPY . .
# ENV GOPROXY="https://goproxy.cn,direct"
# ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go mod tidy
ARG VERSION=unknown
RUN go build -ldflags="-X main.version=${VERSION} -w -s" -a -o dist/ecmdb main.go
FROM chromedp/headless-shell:133.0.6905.0 as deploy
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
ENV EGO_DEBUG=true
ENV PATH /headless-shell:$PATH
WORKDIR /app
COPY --from=build /app/dist/ecmdb .
EXPOSE 8000
ENTRYPOINT ["./ecmdb", "start"]