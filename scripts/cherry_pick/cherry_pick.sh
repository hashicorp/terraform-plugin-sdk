#!/bin/bash

# WIP

COMMIT_ID=$1

# go and get the rev list
REV_LIST=$(cd ../terraform; git rev-list $COMMIT_ID..HEAD)

for c in $REV_LIST
do

echo "Cherry-picking changes..."
git cherry-pick --no-commit "$c"

echo "Removing non-SDK packages..."
git ls-files | grep -Evf ../terraform-plugin-sdk/SDK_PATTERNS | cut -d / -f 1-2 | uniq | ifne xargs -n1 git rm --quiet -rf

echo "Moving changed files to new paths..."
git ls-files --others | xargs -I '{}' mv '{}' sdk/ ;

done
