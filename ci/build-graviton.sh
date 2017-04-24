#!/bin/bash

set -eu

export AWS_ACCESS_KEY_ID=$1
export AWS_SECRET_ACCESS_KEY=$2
GRAV_REPO=$3
OUT_DIR=$4
STAGE_BUCKET=$5
THIS_DIR=$(pwd)

# we know what the docker container looks like
cp -r graviton-repo /usr/local/src/go/src/github.com/stardog-union/stardog-graviton
cd /usr/local/src/go/src/github.com/stardog-union/stardog-graviton
export GOPATH=/usr/local/src/go
export PATH=/usr/local/go/bin:/usr/local/src/go/bin:$PATH
make
make test
VER=$(cat etc/version | sed 's/-.*//')

gox -osarch="linux/amd64" -osarch="darwin/amd64" -output=$THIS_DIR/$OUT_DIR/$VER/{{.OS}}/stardog-graviton

DATE=$(date +%F-%H-%M-%S)
LINUX_BUCKET_NAME=s3://$STAGE_BUCKET/$VER/$DATE/linux/stardog-graviton
OSX_BUCKET_NAME=s3://$STAGE_BUCKET/$VER/$DATE/darwin/stardog-graviton

echo "XXX aws s3 cp $THIS_DIR/$OUT_DIR/$VER/linux/stardog-graviton $LINUX_BUCKET_NAME"
aws s3 cp $THIS_DIR/$OUT_DIR/$VER/linux/stardog-graviton $LINUX_BUCKET_NAME
echo "YYY aws s3 cp $THIS_DIR/$OUT_DIR/$VER/darwin/stardog-graviton $OSX_BUCKET_NAME"
aws s3 cp $THIS_DIR/$OUT_DIR/$VER/darwin/stardog-graviton $OSX_BUCKET_NAME

echo $LINUX_BUCKET_NAME > $THIS_DIR/$OUT_DIR/s3filename

echo $LINUX_BUCKET_NAME
echo $THIS_DIR/$OUT_DIR/s3filename
cat $THIS_DIR/$OUT_DIR/s3filename
