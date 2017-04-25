#!/bin/bash

set -eu

env

TAG=$1
GRAV_REPO=$2
LINUX_STAGE=$3
DARWIN_STAGE=$4
EXE_OUTPUT=$5

echo $TAG

THIS_DIR=$(pwd)
LINUX_GRAV_EXE=$(ls $THIS_DIR/$LINUX_STAGE/stardog-graviton-*)
DARWIN_GRAV_EXE=$(ls $THIS_DIR/$DARWIN_STAGE/stardog-graviton-*)
DESTNAME=$(basename $LINUX_GRAV_EXE)

if [ "X$TAG" != "X" ]; then
    DESTNAME=stardog-graviton-$TAG
    pushd $GRAV_REPO

    git config --global user.name "Release Pipeline"
    git config --global user.email "support@stardog.com"
    echo "${GIT_SSH_KEY}" > /tmp/key
    chmod 600 /tmp/key
    export GIT_SSH_COMMAND="ssh -i /tmp/key -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    git tag $TAG
    git tag
    echo "Push the tags $GIT_SSH_COMMAND"
    git push origin $TAG
    popd
fi

echo "make the dirs"
ls $EXE_OUTPUT
mkdir -p $EXE_OUTPUT/linux
mkdir -p $EXE_OUTPUT/darwin

cp $LINUX_GRAV_EXE $EXE_OUTPUT/linux/$DESTNAME
cp $DARWIN_GRAV_EXE $EXE_OUTPUT/darwin/$DESTNAME
