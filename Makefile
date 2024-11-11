.PHONY: build
build:
	go build ./cmd/jmddns

.PHONY: format
format:
	@gofmt -s -w .
