GOFMT_FILES?=$$(find . -name '*.go')

default: test

test: fmtcheck generate
	go test ./...

generate:
	go generate ./...

# We separate the protobuf generation because most development tasks on
# Terraform do not involve changing protobuf files and protoc is not a
# go-gettable dependency and so getting it installed can be inconvenient.
#
# If you are working on changes to protobuf interfaces you may either use
# this target or run the individual scripts below directly.
protobuf:
	bash scripts/protobuf-check.sh
	bash internal/tfplugin5/generate.sh

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: default fmt fmtcheck generate test
