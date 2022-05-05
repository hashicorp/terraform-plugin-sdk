GOFMT_FILES?=$$(find . -name '*.go')

WEBSITE_REPO=github.com/hashicorp/terraform-website
BRANCH=master
PWD=$$(pwd)
REPO=$$(basename `git rev-parse --show-toplevel`)
DOCKER_IMAGE="hashicorp/terraform-website:full"
DOCKER_IMAGE_LOCAL="hashicorp-terraform-website-local"
DOCKER_RUN_FLAGS=--interactive \
	--rm \
	--tty \
	--workdir "/website" \
	--volume "$(shell pwd)/website:/website/preview" \
	--publish "3000:3000" \
	-e "IS_CONTENT_PREVIEW=true" \
	-e "PREVIEW_FROM_REPO=$(REPO)" \
	-e "NAV_DATA_DIRNAME=./preview/data" \
	-e "CONTENT_DIRNAME=./preview/docs" \
	-e "CURRENT_GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD)"

default: test

test: fmtcheck generate
	go test ./...

generate:
	go generate ./...

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# Run the terraform.io website to preview local content changes 
website:
	@echo "==> Downloading latest Docker image..."
	@docker pull ${DOCKER_IMAGE}
	@echo "==> Starting website in Docker..."
	@docker run ${DOCKER_RUN_FLAGS} ${DOCKER_IMAGE} npm start

# Run the terraform.io website via a locally built docker image
# 
# This is useful for development as you may be building an image
# with local changes from the `terraform-website` repo.
website/local:
	@echo "==> Starting website in Docker..."
	@docker run ${DOCKER_RUN_FLAGS} ${DOCKER_IMAGE_LOCAL} npm start

# Build the terraform.io image using `terraform-website` git repo as the build context
website/build-local:
	@echo "==> Building local Docker image"
	@docker build https://github.com/hashicorp/terraform-website.git\#$(BRANCH) \
		-t $(DOCKER_IMAGE_LOCAL)

.PHONY: default fmt fmtcheck generate test website website/local website/build-local