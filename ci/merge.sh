#!/bin/bash

set -ue

LOCAL_BRANCH=tempmerger$RANDOM
git checkout -b $LOCAL_BRANCH

REMOTE_URL=$(echo $REMOTE_REPO | sed s^://^://${GITHUB_CREDS_USR}:${GITHUB_CREDS_PSW}@^)

echo "Current branch $LOCAL_BRANCH"
git remote add to_push $REMOTE_REPO
git fetch to_push

git checkout to_push/$MERGE_BRANCH -b $MERGE_BRANCH
git merge --ff-only $LOCAL_BRANCH
if [ "X$TAG_VERSION" != "X" ]; then
    git push $REMOTE_URL $TAG_VERSION
fi
git push $REMOTE_URL $MERGE_BRANCH
