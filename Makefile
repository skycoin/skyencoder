.DEFAULT_GOAL := help
.PHONY: build test bench generate-tests check-generate-tests
.PHONY: generate-benchmark-encoder check-generate-benchmark-encoder-unchanged
.PHONY: format help

build: ## Build skyencoder binary
	go build cmd/skyencoder/skyencoder.go

test: ## Run tests
	go test ./...

check: generate-tests check-generate-tests-unchanged test generate-benchmark-encoder check-generate-benchmark-encoder-unchanged ## Run tests and check code generation

bench: ## Run benchmarks
	go test -benchmem -bench '.*' ./benchmark

generate-tests: ## Generate encoders and test for test objects
	go run cmd/skyencoder/skyencoder.go -struct DemoStruct -output-file autogen_demo_struct_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct DemoStructOmitEmpty -output-file autogen_demo_struct_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenStringStruct1 -output-file autogen_max_len_string_struct1_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenStringStruct2 -output-file autogen_max_len_string_struct2_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenAllStruct1 -output-file autogen_max_len_all_struct1_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenAllStruct2 -output-file autogen_max_len_all_struct2_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedSliceStruct1 -output-file autogen_max_len_nested_slice_struct1_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedSliceStruct2 -output-file autogen_max_len_nested_slice_struct2_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapKeyStruct1 -output-file autogen_max_len_nested_map_key_struct1_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapKeyStruct2 -output-file autogen_max_len_nested_map_key_struct2_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapValueStruct1 -output-file autogen_max_len_nested_map_value_struct1_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapValueStruct2 -output-file autogen_max_len_nested_map_value_struct2_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct OnlyOmitEmptyStruct -output-file autogen_only_omit_empty_struct_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyStruct -output-file autogen_omit_empty_struct_skyencoder_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyMaxLenStruct1 -output-file autogen_omit_empty_max_len_struct1_test.go github.com/skycoin/skyencoder
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyMaxLenStruct2 -output-file autogen_omit_empty_max_len_struct2_test.go github.com/skycoin/skyencoder

check-generate-tests: ## Check that make generate-tests did not change the code
	@if [ "$(shell git diff ./autogen_demo_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_demo_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_demo_struct_omit_empty_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_demo_struct_omit_empty_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_string_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_string_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_string_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_string_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_all_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_all_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_all_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_all_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_slice_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_slice_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_slice_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_slice_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_key_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_key_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_key_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_key_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_value_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_value_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_value_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_max_len_nested_map_value_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_only_omit_empty_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_only_omit_empty_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_max_len_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_max_len_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_max_len_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./autogen_omit_empty_max_len_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi

generate-benchmark-encoder: ## Generate the encoders for the benchmarks
	go run cmd/skyencoder/skyencoder.go -struct BenchmarkStruct github.com/skycoin/skyencoder/benchmark
	go run cmd/skyencoder/skyencoder.go -struct SignedBlock -package benchmark -output-path ./benchmark github.com/skycoin/skycoin/src/coin

check-generate-benchmark-encoder-unchanged: ## Check that make generate-benchmark-encoder did not change the code
	@if [ "$(shell git diff ./benchmark/benchmark_struct_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/benchmark_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/signed_block_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/signed_block_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmark-encoder' ; exit 2 ; fi

format:  # Formats the code. Must have goimports installed (use make install-linters).
	# This sorts imports
	goimports -w .
	# This performs code simplifications
	gofmt -s -w .

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
