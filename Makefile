ROLE := adapter-agentic
.PHONY: build test
build:
	go build -o bin/cofiswarm-adapter-agentic ./cmd/cofiswarm-adapter-agentic
test: build test-standalone-layout test-adapter-config-gate
test-standalone-layout:
	./test/scripts/assert-layout.sh $(ROLE)
test-adapter-config-gate:
	./test/scripts/test-adapter-config-gate.sh $(ROLE)
