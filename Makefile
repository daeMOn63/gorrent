
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: test ## Build the binary
	go build -o build/gorrent main.go

.PHONY: test
test: ## Run tests
	go test -coverprofile=/tmp/go-code-cover -timeout 30s  ./...

.PHONY: cover
cover: test ## Show coverage
	go tool cover -html=/tmp/go-code-cover

.PHONY: clean
clean:
	rm -rf build/
