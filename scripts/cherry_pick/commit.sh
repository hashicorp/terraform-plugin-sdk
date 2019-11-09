#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: ./scripts/cherry_pick/commit.sh MERGE_COMMIT_HASH"
fi

function pleaseUseGNUsed {
    echo "Please install GNU sed to your PATH as 'sed'."
    exit 1
}
sed --version > /dev/null || pleaseUseGNUsed

COMMIT_ID=$1

COMMIT_MSG=$(git log --format=%B -n 1 "$COMMIT_ID" | \
                 sed -z -r 's/Merge pull request (#[0-9]+) from ([^\n]*\/[^\n]*)\n\n(.*$)/\3\nThis commit was generated from hashicorp\/terraform\1./g')

git commit -C "$COMMIT_ID" && \
    # amend commit message afterwards to preserve authorship information
    git commit --amend --message "${COMMIT_MSG}"
