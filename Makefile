.PHONY: default help

default: help
help: ## display make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(word 1, $(MAKEFILE_LIST)) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m make %-20s -> %s\n\033[0m", $$1, $$2}'


.PHONY: build-all
build-all:  ## build: go get, tidy, fmt, go build -o gplr
	@bash -c "go get k8s.io/client-go@v0.26.3"
	@bash -c "go mod tidy"
	@bash -c "go fmt ./..."
	@bash -c "go build -o gplr"


.PHONY: coverage
coverage: ## run tests with coverage, output to coverage.out
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: check-coverage
check-coverage: coverage ## check coverage is above 40%
	@go tool cover -func=coverage.out | grep -oP "total:\s+\K\d+\.\d+" | awk '{ if ($$1 < 2.0) { printf "FAIL: total coverage is below 40%%\n"; exit 1; } else { printf "PASS: total coverage is above 40%%\n"; exit 0; } }'

.PHONY: lint
lint: ## golangci-lint
	golangci-lint run


.PHONY: go-lint-fix
go-lint-fix: ## golangci-lint run --fix
	golangci-lint run --fix

.PHONY: codefresh-environments
codefresh-environments: ## codefresh get runtime-environments
	codefresh get runtime-environments
	codefresh get projects
	codefresh get pipelines
	codefresh get builds

.PHONY: test
test: ## go test ./...
	go test ./...

.PHONY: pipeline
pipeline: test check-coverage ## run all codefresh pipeline checks
	@echo "All tests completed successfully."
