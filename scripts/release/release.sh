#!/bin/bash

set -e

# Prepare will prepare the repository for release
# 1. Modify changelog
# 2. Run changelog links script
# 3. Commit changes
# 4. Create a Git tag

function pleaseUseGNUsed {
    echo "Please install GNU sed to your PATH as 'sed'."
    exit 1
}

function init {
  sed --version > /dev/null || pleaseUseGNUsed

  DATE=`date '+%B %d, %Y'`
  START_DIR=`pwd`

  if [ "$CI" = true ] ; then
    GPG_KEY_ID=FBA4BCB482BC44B5
    gpg --import <(echo -e "${GPG_PUBLIC_KEY}")
    gpg --import <(echo -e "${GPG_PRIVATE_KEY}")
    git config --global user.email hashibot-feedback+tf-sdk-circleci@hashicorp.com
    git config --global user.name "Terraform SDK CircleCI"
  fi

  TARGET_VERSION="$(getTargetVersion)"
}

function getTargetVersion {
  # parse target version from CHANGELOG
  sed -n 's/^# \([0-9]\+\.[0-9]\+\.[0-9]\+\) (Unreleased)$/\1/p' CHANGELOG.md || \
     (echo "\nTarget version not found in changelog, exiting" && \
       exit 1)
}

function modifyChangelog {
  sed -i "s/$TARGET_VERSION (Unreleased)$/$TARGET_VERSION ($DATE)/" CHANGELOG.md
}

function changelogLinks {
  ./scripts/release/changelog_links.sh
}

function changelogMain {
  printf "Modifying Changelog..."
  modifyChangelog
  printf "ok!\n"
  printf "Running Changelog Links..."
  changelogLinks
  printf "ok!\n"
}

function modifyVersionFiles {
  sed -i "s/var SDKVersion =.*/var SDKVersion = \"${TARGET_VERSION}\"/" meta/meta.go
  sed -i "s/var SDKPrerelease =.*/var SDKPrerelease = \"\"/" meta/meta.go
}

function commitChanges {
  git add CHANGELOG.md
  modifyVersionFiles
  git add meta/meta.go

  if [ "$CI" = true ] ; then
      git commit --gpg-sign="${GPG_KEY_ID}" -m "v${TARGET_VERSION} [skip ci]"
      git tag -a -m "v${TARGET_VERSION}" -s -u "${GPG_KEY_ID}" "v${TARGET_VERSION}"
  else
      git commit -m "v${TARGET_VERSION} [skip ci]"
      git tag -a -m "v${TARGET_VERSION}" -s "v${TARGET_VERSION}"
  fi

  git push origin master
  git push origin "v${TARGET_VERSION}"
}

function commitMain {
  printf "Committing Changes..."
  commitChanges
  printf "ok!\n"
}

function main {
  init
  changelogMain
  commitMain
}

main
