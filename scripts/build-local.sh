#!/bin/bash


THIS_DIR=`dirname $0`
cd $THIS_DIR/..
BASE_DIR=`pwd`
BIN_DIR=$BASE_DIR/bin

v=$(git describe --abbrev=0 --tags 2> /dev/null)
if [ $? -ne 0 ]; then
    v="unknown"
fi
which go-bindata
if [ $? -ne 0 ]; then
    go get -u github.com/jteeuwen/go-bindata/...
fi
set -e
t=$(git rev-parse HEAD)
echo $v-$t > etc/version

export PATH=$GOPATH/bin:$PATH

go-bindata -prefix aws -o aws/data.go -pkg aws aws/etc/...
go-bindata -o data.go -pkg main etc/...

go install github.com/stardog-union/stardog-graviton
