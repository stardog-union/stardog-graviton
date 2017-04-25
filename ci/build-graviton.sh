#!/bin/bash

set -eu

GRAV_REPO=$1
OUT_DIR=$2
THIS_DIR=$(pwd)

# we know what the docker container looks like
cp -r graviton-repo /usr/local/src/go/src/github.com/stardog-union/stardog-graviton
cd /usr/local/src/go/src/github.com/stardog-union/stardog-graviton
export GOPATH=/usr/local/src/go
export PATH=/usr/local/go/bin:/usr/local/src/go/bin:$PATH
make
make test
VER=$(cat etc/version)

echo $VER

gox -osarch="linux/amd64" -osarch="darwin/amd64" -output=$THIS_DIR/$OUT_DIR/{{.OS}}/stardog-graviton-$VER

ls -l $THIS_DIR/$OUT_DIR/darwin
ls -l $THIS_DIR/$OUT_DIR/linux
