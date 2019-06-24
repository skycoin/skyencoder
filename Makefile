.DEFAULT_GOAL := help
.PHONY: build test test-386 test-amd64 bench generate
.PHONY: check-generate-unchanged
.PHONY: generate-tests check-generate-tests-unchanged
.PHONY: generate-benchmarks check-generate-benchmarks-unchanged
.PHONY: format help

build: ## Build skyencoder binary
	go build cmd/skyencoder/skyencoder.go

test: ## Run tests
	go test ./...

test-386: ## Run tests on 386 arch
	CGO_ENABLED=0 GOARCH=386 go test ./...

test-amd64: ## Run tests on 386 arch
	CGO_ENABLED=0 GOARCH=amd64 go test ./...

check: generate check-generate-unchanged test-386 test-amd64 ## Run tests and check code generation

bench: ## Run benchmarks
	go test -benchmem -bench '.*' ./benchmark

generate: generate-tests generate-benchmarks ## Generate all test and benchmarks

check-generate-unchanged: check-generate-tests-unchanged check-generate-benchmarks-unchanged

generate-tests: ## Generate encoders and test for test objects
	go run cmd/skyencoder/skyencoder.go -struct DemoStruct -output-file demo_struct_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct DemoStructOmitEmpty -output-file demo_struct_omit_empty_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct DemoStructNestedBytes -output-file demo_struct_nested_bytes_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenStringStruct1 -output-file max_len_string_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenStringStruct2 -output-file max_len_string_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenAllStruct1 -output-file max_len_all_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenAllStruct2 -output-file max_len_all_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedSliceStruct1 -output-file max_len_nested_slice_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedSliceStruct2 -output-file max_len_nested_slice_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapKeyStruct1 -output-file max_len_nested_map_key_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapKeyStruct2 -output-file max_len_nested_map_key_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapValueStruct1 -output-file max_len_nested_map_value_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct MaxLenNestedMapValueStruct2 -output-file max_len_nested_map_value_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct OnlyOmitEmptyStruct -output-file only_omit_empty_struct_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyStruct -output-file omit_empty_struct_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyMaxLenStruct1 -output-file omit_empty_max_len_struct1_skyencoder_test.go github.com/skycoin/skyencoder/tests
	go run cmd/skyencoder/skyencoder.go -struct OmitEmptyMaxLenStruct2 -output-file omit_empty_max_len_struct2_skyencoder_test.go github.com/skycoin/skyencoder/tests

check-generate-tests-unchanged: ## Check that make generate-tests did not change the code
	@if [ "$(shell git diff ./tests/demo_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/demo_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/demo_struct_omit_empty_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/demo_struct_omit_empty_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/demo_struct_nested_bytes_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/demo_struct_nested_bytes_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_string_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_string_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_string_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_string_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_all_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_all_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_all_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_all_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_slice_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_slice_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_slice_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_slice_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_key_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_key_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_key_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_key_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_value_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_value_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_value_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/max_len_nested_map_value_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/only_omit_empty_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/only_omit_empty_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_struct_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_max_len_struct1_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_max_len_struct1_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_max_len_struct2_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi
	@if [ "$(shell git diff ./tests/omit_empty_max_len_struct2_skyencoder_test_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-tests' ; exit 2 ; fi

generate-benchmarks: ## Generate the encoders for the benchmarks
	go run cmd/skyencoder/skyencoder.go -struct BenchmarkStruct github.com/skycoin/skyencoder/benchmark
	go run cmd/skyencoder/skyencoder.go -struct SignedBlock -package benchmark -output-path ./benchmark github.com/skycoin/skycoin/src/coin

check-generate-benchmarks-unchanged: ## Check that make generate-benchmarks did not change the code
	@if [ "$(shell git diff ./benchmark/benchmark_struct_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmarks' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/benchmark_struct_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmarks' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/signed_block_skyencoder.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmarks' ; exit 2 ; fi
	@if [ "$(shell git diff ./benchmark/signed_block_skyencoder_test.go | wc -l | tr -d ' ')" != "0" ] ; then echo 'Changes detected after make generate-benchmarks' ; exit 2 ; fi

format:  ## Formats the code. Must have goimports installed (use make install-linters).
	# This sorts imports
	goimports -w .
	# This performs code simplifications
	gofmt -s -w .

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
