GOFMT_FILES?=$$(find . -name '*.go')

default: test

test: fmtcheck generate
	go test ./...

generate:
	go generate ./...

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: default fmt fmtcheck generate test
