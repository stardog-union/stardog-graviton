#!/bin/bash

set -u

TAG=$1
GRAV_REPO=$2
LINUX_STAGE=$3
DARWIN_STAGE=$4
EXE_OUTPUT=$5

THIS_DIR=$(pwd)
LINUX_GRAV_EXE=$(ls $THIS_DIR/$LINUX_STAGE/stardog-graviton-*)
DARWIN_GRAV_EXE=$(ls $THIS_DIR/$DARWIN_STAGE/stardog-graviton-*)
DESTNAME=$(basename $LINUX_GRAV_EXE)

VER=$(echo $DESTNAME | sed 's/stardog-graviton-//')
echo "TAG: $TAG"
echo "VER: $VER"
$LINUX_GRAV_EXE --version
$LINUX_GRAV_EXE --version  2>&1 | grep $VER
if [ $? -ne 0 ]; then
    echo "The version information is not correct in the artifact"
    exit 1
fi

set -e
if [ "X$TAG" != "X0" ]; then
    set +e
    $LINUX_GRAV_EXE --version  2>&1 | grep $TAG
    if [ $? -ne 0 ]; then
        echo "The version information is not correct in the artifact and tag"
        exit 1
    fi
    set -e
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
    VER=$TAG
    popd
fi

echo "make the dirs"
ls $EXE_OUTPUT
mkdir -p $EXE_OUTPUT/$VER

echo "Create the linux zip"
cp $LINUX_GRAV_EXE stardog-graviton
chmod 755 stardog-graviton
zip "$EXE_OUTPUT/$VER/stardog-graviton_$VER""_linux_amd64.zip" stardog-graviton
echo "Create the darwin zip"
cp $DARWIN_GRAV_EXE stardog-graviton
chmod 755 stardog-graviton
zip "$EXE_OUTPUT/$VER/stardog-graviton_$VER""_darwin_amd64.zip" stardog-graviton
ls $EXE_OUTPUT/$VER
