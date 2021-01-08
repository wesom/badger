GOPATH:=$(shell go env GOPATH)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
BIN:=snakepit

.PHONY: build test

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -v -o $(BIN)

test:
	go test -coverpkg ./...

clean:
	rm $(BIN)
