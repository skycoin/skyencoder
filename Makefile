.DEFAULT_GOAL := help
.PHONY: build test bench generate-benchmark-encoder check-generate-benchmark-encoder-unchanged format help

build: ## Build skyencoder binary
	go build cmd/skyencoder/skyencoder.go

test: ## Run tests
	go test ./...

check: test generate-benchmark-encoder check-generate-benchmark-encoder-unchanged ## Run tests and check code generation

bench: ## Run benchmarks
	go test -benchmem -bench '.*' ./benchmark

generate-benchmark-encoder: ## Generate the encoders for the benchmarks
	go run cmd/skyencoder/skyencoder.go -struct BenchmarkStruct github.com/skycoin/skyencoder/benchmark
	go run cmd/skyencoder/skyencoder.go -struct SignedBlock -package benchmark -output-path ./benchmark github.com/skycoin/skycoin/src/coin

check-generate-benchmark-encoder-unchanged: ## Check that make generate-benchmark-encoder did not change the code
	@if [ "$(shell git diff ./benchmark/benchmark_struct_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/signed_block_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi

format:  # Formats the code. Must have goimports installed (use make install-linters).
	# This sorts imports
	goimports -w .
	# This performs code simplifications
	gofmt -s -w .

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
