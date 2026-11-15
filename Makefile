.PHONY: all build test lint benchmark clean

all: build test

build:
	go build -o repojet .

test:
	go test -v -count=1 ./...

lint:
	go vet ./...
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; running go vet only"; \
	fi

benchmark:
	go run ./benchmark

clean:
	rm -f repojet
