#!/bin/bash
set -e

SDK_FOLDERS="helper/acctest
helper/customdiff
helper/encryption
helper/hashcode
helper/logging
helper/mutexkv
helper/pathorcontents
helper/resource
helper/schema
helper/structure
helper/validation
httpclient
plugin
terraform
"

echo "Finding imports ..."
IMPORTS=$(echo "$SDK_FOLDERS" | xargs -I{} go list -json ./{} | jq -r ".ImportPath" | sort | uniq | sed -e 's/^github.com\/hashicorp\/terraform\///')
echo "Finding build dependencies ..."
DEPS=$(echo "$SDK_FOLDERS" | xargs -I{} go list -json ./{} | jq -r ". | select((.Deps | length) > 0) | (.Deps[] + \"\n\" + .ImportPath) | select(startswith(\"github.com/hashicorp/terraform/\"))" | sort | uniq | sed -e 's/^github.com\/hashicorp\/terraform\///')

echo "Finding test dependencies ..."
TEST_IMPORTS=$(echo "$DEPS" | xargs -I{} go list -json ./{} | jq -r ". | select((.TestImports | length) > 0) | .TestImports[] | select(startswith(\"github.com/hashicorp/terraform/\"))" | sort | uniq | sed -e 's/^github.com\/hashicorp\/terraform\///')
TEST_DEPS=$(echo "$TEST_IMPORTS" | xargs -I{} go list -json ./{} | jq -r ". | select((.Deps | length) > 0) | (.Deps[] + \"\n\" + .ImportPath) | select(startswith(\"github.com/hashicorp/terraform/\"))" | sort | uniq | sed -e 's/^github.com\/hashicorp\/terraform\///')

echo "All dependencies found."

mkdir -p ../terraform-plugin-sdk/sdk/internal

# Find all SDK related files
ALL_PKGS=$(printf "${IMPORTS}\n${DEPS}\n${TEST_IMPORTS}\n${TEST_DEPS}" | sort | uniq)
COUNT_PKG=$(echo "$ALL_PKGS" | wc -l | tr -d ' ')
echo "Finding files of ${COUNT_PKG} packages ..."

SDK_FILES=$(echo "$ALL_PKGS" | xargs -I_  find . -path ./_/* \( -path './_/testdata*' -or -prune \))
SDK_LIST_PATH=$(mktemp)

echo "$SDK_FILES" > $SDK_LIST_PATH
echo "SDK files listed in ${SDK_LIST_PATH}"

# echo "Computing the difference ..."
# TO_REMOVE_PATH=$(mktemp)
# NONSDK_FILES=$(find . -type f | grep -Fxvf $SDK_LIST_PATH | grep -v  > $TO_REMOVE_PATH)
# rm -f $SDK_LIST_PATH
# echo "Files to remove are listed in ${TO_REMOVE_PATH}"

# Copy SDK files to new place
# TODO: git filter-branch --index-filter 'echo "$TO_REMOVE_PATH" | xargs -I{} git rm --cached --ignore-unmatch {}' HEAD
cat $SDK_LIST_PATH | xargs -I{} rsync -m -lptgoDd -R {} ../terraform-plugin-sdk/sdk/

cd ../terraform-plugin-sdk/sdk

# Change import paths
echo "Changing import paths from terraform to terraform-plugin-sdk ..."
find . -name '*.go' | xargs -I{} sed -i 's/github.com\/hashicorp\/terraform\([\/"]\)/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\1/' {}

echo "Initializing go modules ..."
cd ..
go mod init
go get ./...
go mod tidy
echo "Go modules initialized."

# TODO: Check version of grep to make sure it's GNU 3.3+

echo "Moving internal packages up ..."
# Flatten sdk/internal/* into sdk/* to avoid nested internal packages & breaking import trees
INTERNAL_FOLDERS=$(go list -json ./... | jq -r .ImportPath | sed -e 's/^github.com\/hashicorp\/terraform-plugin-sdk\/sdk\///' | ggrep -E '^internal\/' | sed -e 's/^internal\///')
cd ./sdk
echo "$INTERNAL_FOLDERS" | xargs -I{} mv ./internal/{} ./{}
rm -rf ./internal
echo "$INTERNAL_FOLDERS" | sed 's/\//\\\\\//g' | xargs -I{} sh -c "find . -name '*.go' | xargs -I@ sed -i 's/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\/internal\/{}/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\/{}/' @"
echo "Internal packages moved."

echo "Finding non-SDK packages & folders ..."
# Internalize non-SDK packages
SDK_PKGS_LIST_PATH=$(mktemp)
# Find all parent folders first
PARENT_FOLDERS="$SDK_FOLDERS"
while [[ $(echo -n "$PARENT_FOLDERS" | wc -l) -gt 0 ]]; do
	PARENT_FOLDERS=$(echo "$PARENT_FOLDERS" | xargs -I{} dirname {} | ggrep -xFv '.' | sort | uniq)
	echo "$PARENT_FOLDERS" | xargs -I{} echo ./{} > $SDK_PKGS_LIST_PATH
done

echo "$SDK_FOLDERS" | xargs -I{} echo ./{} >> $SDK_PKGS_LIST_PATH

echo "SDK packages stored in $SDK_PKGS_LIST_PATH"

SDK_FOLDERS_PATTERNS_PATH=$(mktemp)
cat $SDK_PKGS_LIST_PATH | xargs -I{} sh -c "echo ^{}\$; echo ^{}/testdata" > $SDK_FOLDERS_PATTERNS_PATH
NONSDK_FOLDERS=$(find . -type d -and \( ! -path './.git*' \) | ggrep -xFv '.' | ggrep -v -f $SDK_FOLDERS_PATTERNS_PATH)
NONSDK_GO_PKGS=$(go list -json ./... | jq -r .ImportPath | sed -e 's/^github.com\/hashicorp\/terraform-plugin-sdk\/sdk/\./' | ggrep -xFv -f $SDK_PKGS_LIST_PATH | sed -e 's/^\.\///')

NONSDK_GO_PKGS_PATH=$(mktemp)
echo "$NONSDK_GO_PKGS" > $NONSDK_GO_PKGS_PATH
echo "NonSDK packages stored in $NONSDK_GO_PKGS_PATH"

NONSDK_FOLDERS_PATH=$(mktemp)
echo "$NONSDK_FOLDERS" > $NONSDK_FOLDERS_PATH
echo "NonSDK folders stored in $NONSDK_FOLDERS_PATH"

# Move all non-SDK folders
echo "Moving non-SDK folders under internal ..."
echo "$NONSDK_FOLDERS" | xargs -I{} sh -c 'mkdir -p $(dirname ./internal/{}); [ -d {} ] && mv -v {} ./internal/{} || true'
echo "Non-SDK folders moved."

# Fix imports in newly moved non-SDK packages
COUNT_NONSDK_GO_PKGS=$(echo "$NONSDK_GO_PKGS" | wc -l | tr -d ' ')
echo "Updating $COUNT_NONSDK_GO_PKGS import paths for moved files ..."
echo "$NONSDK_GO_PKGS" | sed 's/\//\\\\\//g' | xargs -I{} sh -c "find . -name '*.go' | xargs -I@ sed -i 's/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\/{}\([\/\"]\)/github.com\/hashicorp\/terraform-plugin-sdk\/sdk\/internal\/{}\1/' @"
echo "Import paths updated."
