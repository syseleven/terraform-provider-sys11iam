.PHONY: format unit-test

terraform-provider-sys11iam:
	go build -ldflags "-X github.com/syseleven/terraform-provider-sys11iam/tmp_main.Version=$(shell git describe --tags --always)"

tf-generate:
	tfplugingen-openapi generate --config ./generator_config.yml --output ./provider-code-spec.json ./openapi.json
	tfplugingen-framework generate resources --input ./provider-code-spec.json --output ./internal

format:
	go fmt ./...
	find . -name '*.go' -exec sed -i '/import (/,/)/{ /^[ \t]*$$/d}' {} \;
	goimports -w .

unit-test:
	go generate ./...
	gotestsum --format testname ./... -p 1 -v

unit-test-ci:
	go generate ./...
	gotestsum --junitfile test-report.xml --format testname ./... -p 1 -v -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep total:

unit-test-cov:
	go generate ./...
	gotestsum -- -p 1 -v -coverprofile=coverage.out -covermode count ./...
	gocover-cobertura < coverage.out > cov.xml
