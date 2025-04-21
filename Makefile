GOFMT_FILES?=$$(find . -name '*.go')

default: test

test: generate
	go test ./...

lint:
	golangci-lint run

generate:
	go generate ./...
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e $(GOFMT_FILES)

# Run this if working on the website locally to run in watch mode.
website:
	$(MAKE) -C website website

# Use this if you have run `website/build-local` to use the locally built image.
website/local:
	$(MAKE) -C website website/local

# Run this to generate a new local Docker image.
website/build-local:
	$(MAKE) -C website website/build-local

.PHONY: default fmt lint generate test website website/local website/build-local
