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


