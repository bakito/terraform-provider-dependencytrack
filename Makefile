# Include toolbox tasks
include ./.toolbox.mk

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

install:
	go install .

generate: install
	rm -Rf docs
	go generate ./...

release: tb.goreleaser tb.semver
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean

