.PHONY: build test test-e2e lint check examples clean

build:
	go build ./...

test:
	go test ./...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

lint:
	golangci-lint run

check:
	@test -z "$$(gofmt -l .)" || { echo "not gofmt-formatted:"; gofmt -l .; exit 1; }
	go vet ./...

examples:
	for d in examples/*/; do go build "./$$d"; done

clean:
	rm -rf bin/
