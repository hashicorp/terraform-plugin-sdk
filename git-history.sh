#!/bin/bash
set -e

git filter-branch -f --index-filter 'for file in $(cat /home/ec2-user/SDK_TO_REMOVE); do git rm -rf --cached --ignore-unmatch ${file}; done' HEAD

