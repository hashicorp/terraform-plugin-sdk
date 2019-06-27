#!/bin/bash
set -e

find . \( -name '*.go' -or -name '*.sh' \) -or \( -path './vendor' -prune \) | grep -v vendor | \
	xargs -I{} sed -i 's/github.com\/hashicorp\/terraform\([\/"]\)/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\1/g' {}

go get ./...
go mod tidy
go mod vendor
