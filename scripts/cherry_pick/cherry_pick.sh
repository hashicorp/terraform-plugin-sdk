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

echo "Unstaging files where SDK intentionally diverges from Core..."
for f in $(git diff --name-only --cached | grep -Ef ./scripts/cherry_pick/IGNORE_FILES)
do
    git reset "$f"
    git checkout -- "$f"
done

COMMIT_MSG=$(git log --format=%B -n 1 "$COMMIT_ID" | sed -z -r 's/Merge pull request (#[0-9]+) from ([^\n]*\/[^\n]*)\n\n(.*$)/\3\nThis commit was generated from hashicorp\/terraform\1./g')

echo "Committing changes. If this fails, you must resolve the merge conflict manually."
git commit -C "$COMMIT_ID" && \
# amend commit message afterwards to preserve authorship information
git commit --amend --message "${COMMIT_MSG}" \
&& echo "Success!"
