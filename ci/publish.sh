#!/bin/bash

set -u

START_DIR=$(pwd)
OUTPUT_DIR=${START_DIR}/OUTPUT

if [ "X$TAG_VERSION" != "X" ]; then
    VER=$TAG_VERSION
else
    VER=$(cat etc/version)
fi

set -e 
LINUX_BASE_NAME="stardog-graviton_$VER""_linux_amd64.zip"
LINUX_ZIP=$OUTPUT_DIR/$VER/$LINUX_BASE_NAME

DARWIN_BASE_NAME="stardog-graviton_$VER""_darwin_amd64.zip"
DARWIN_ZIP=$OUTPUT_DIR/$VER/$DARWIN_BASE_NAME

echo "aws --debug --region us-east-1 s3 cp $LINUX_ZIP s3://$S3_BUCKET/$LINUX_BASE_NAME"

cat /etc/issue
aws --version

s3cmd put $LINUX_ZIP s3://$S3_BUCKET/$LINUX_BASE_NAME
aws --region us-east-1 s3 cp $LINUX_ZIP s3://$S3_BUCKET/$LINUX_BASE_NAME
aws --region us-east-1 s3 cp $DARWIN_ZIP s3://$S3_BUCKET/$DARWIN_BASE_NAME
