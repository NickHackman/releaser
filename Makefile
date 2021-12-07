BIN=tagger
SRC=$(shell find . -name "*.go")

.PHONY: build help test fmt lint clean run test-cover

default: help

help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo
	@echo "Targets:"
	@echo "  build                    build binary ${BIN}"
	@echo "  run                      build and run ${BIN}"
	@echo "  lint                     run go vet and golangci-lint"
	@echo "  test                     run unit tests"
	@echo "  test-cover               run tests and view code coverage"
	@echo "  clean                    remove binary"
	@echo "  fmt                      format files using gofmt"

build:
	go get -v -d ./...
	go build -o $(BIN) main.go

run: build
	./$(BIN)

fmt:
	gofmt -s -w $(SRC)

lint:
	go vet ./...
	golangci-lint run

test:
	go test -v -race ./...

test-cover:
	go test ./... -race -coverprofile=coverage.out && go tool cover -html=coverage.out

clean:
	rm -f $(BIN)