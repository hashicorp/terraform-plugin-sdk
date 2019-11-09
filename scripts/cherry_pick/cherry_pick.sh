#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: ./scripts/cherry_pick/cherry_pick.sh MERGE_COMMIT_HASH"
fi

function pleaseUseGNUsed {
    echo "Please install GNU sed to your PATH as 'sed'."
    exit 1
}
sed --version > /dev/null || pleaseUseGNUsed

COMMIT_ID=$1

echo "Cherry-picking changes..."
git cherry-pick --no-commit --mainline 1 "$COMMIT_ID"

echo "Unstaging files removed by us..."
git status --short | sed -n 's/^DU //p' | ifne xargs git rm

echo "Unstaging added files that match the ignore list..."
git status --short | sed -n 's/^A  //p' | grep -Ef ./scripts/cherry_pick/IGNORE_FILES | ifne xargs git rm -f

echo "Unstaging files where SDK intentionally diverges from Core..."
for f in $(git diff --name-only --cached | grep -Ef ./scripts/cherry_pick/IGNORE_FILES)
do
    git reset "$f"
    git checkout -- "$f"
done

echo "Committing changes. If this fails, you must resolve the merge conflict manually."
./scripts/cherry_pick/commit.sh $COMMIT_ID && echo "Success!"
