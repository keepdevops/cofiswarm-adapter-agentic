ROLE := adapter-agentic
.PHONY: build test lint
build:
	go build -o bin/cofiswarm-adapter-agentic ./cmd/cofiswarm-adapter-agentic
lint:
	@test -z "$$(gofmt -l . )" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
	go vet ./...
test: lint build go-test test-standalone-layout test-adapter-config-gate
go-test:
	go test ./...
test-standalone-layout:
	./test/scripts/assert-layout.sh $(ROLE)
test-adapter-config-gate:
	./test/scripts/test-adapter-config-gate.sh $(ROLE)
