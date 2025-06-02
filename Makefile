BINARY_NAME=bet_test_app

.PHONY: all build run clean test deps

all: build

build:
	@go build -o $(BINARY_NAME) ./cmd/app/main.go

run: build
	@./$(BINARY_NAME)

run-with-url:
	@go run ./cmd/app/main.go -apiBURL=https://webhook.site/your-unique-id

clean:
	@go clean
	@rm -f $(BINARY_NAME)

test:
	@go test ./internal/service/...

deps:
	@go mod tidy