#!/bin/bash

mkdir OUTPUT
set -eu

if [ "X$TAG_VERSION" != "X" ]; then
    echo "TAGGING with $TAG_VERSION"
    git tag $TAG_VERSION
    VER=$TAG_VERSION
else
    VER=$(cat etc/version)
fi
make clean
make

cat etc/version

cd /usr/local/src/go/src/github.com/stardog-union/stardog-graviton
echo "Run cross compile..."
gox -osarch="linux/amd64" -osarch="darwin/amd64" -output=OUTPUT/{{.OS}}/stardog-graviton-$VER 
