.PHONY: build run test clean docker-build

APP_NAME := crosscast
BIN_DIR := bin

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/crosscast

run:
	go run ./cmd/crosscast

test:
	go test -v -race ./...

clean:
	rm -rf $(BIN_DIR)

docker-build:
	docker build -t crosscast:latest .
