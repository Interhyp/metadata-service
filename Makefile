.PHONY: generate
generate:
	@./api-generator/generate.sh

.PHONY: test
test:
	@go test ./...