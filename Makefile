.PHONY: build lint test test-integration test-all clean

build:
	go build -o bastion ./cmd/bastion

lint:
	golangci-lint run ./...

test:
	go test ./...

test-integration:
	go test -tags integration ./...

test-all: test test-integration

clean:
	rm -f bastion
