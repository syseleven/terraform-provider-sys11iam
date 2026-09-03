.PHONY: format unit-test tf-generate tf-generate-check

terraform-provider-sys11iam:
	go build -ldflags "-X github.com/syseleven/terraform-provider-sys11iam/tmp_main.Version=$(shell git describe --tags --always)"

dev:
	$(MAKE) terraform-provider-sys11iam
	go install

tf-generate:
	tfplugingen-openapi generate --config ./generator_config.yml --output ./provider-code-spec.json ./openapi.json
	tfplugingen-framework generate resources --input ./provider-code-spec.json --output ./internal
	./scripts/post-generate.sh

tf-generate-check:
	@echo "Checking that generated code is up-to-date..."
	$(MAKE) tf-generate
	@if ! git diff --quiet internal/resource_*/ provider-code-spec.json; then \
		echo "ERROR: Generated code is out of sync. Run 'make tf-generate' and commit the result."; \
		git --no-pager diff --stat internal/resource_*/ provider-code-spec.json; \
		exit 1; \
	fi
	@echo "Generated code is up-to-date."

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

docs:
	tfplugindocs generate
