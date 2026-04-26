.PHONY: build test test-e2e lint examples clean

build:
	go build ./...

test:
	go test ./...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

lint:
	golangci-lint run

examples:
	for d in examples/*/; do go build "./$$d"; done

clean:
	rm -rf bin/
