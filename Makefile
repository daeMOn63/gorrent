
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	go build -o build/gorrent main.go

.PHONY: create
test: ## Run tests
	go test -coverprofile=/tmp/go-code-cover -timeout 30s  ./...

.PHONY: clean
clean:
	rm -rf build/
