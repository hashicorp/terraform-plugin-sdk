#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: ./scripts/cherry_pick/cherry_pick.sh MERGE_COMMIT_HASH"
fi

COMMIT_ID=$1

echo "Cherry-picking changes..."
git cherry-pick --no-commit --mainline 1 "$COMMIT_ID"

echo "Unstaging files removed by us..."
git status --short | sed -n 's/^DU //p' | ifne xargs git rm

echo "Committing changes. If this fails, you must resolve the merge conflict manually."
git commit -C "$COMMIT_ID" && echo "Success!"
