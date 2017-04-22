#!/bin/bash

set -e

SKIP=$1
export AWS_ACCESS_KEY_ID=$2
export AWS_SECRET_ACCESS_KEY=$3

STAGE_BUCKET=$4
BUILD_DIR=$5
OLDDIR=$6

if [ $SKIP -eq 1 ]; then
    echo "Skipping the graviton tests"
    exit 0
fi

ls
echo "BUILD"
ls $BUILD_DIR

echo "OLD"
ls $OLDDIR

THIS_DIR=$(pwd)
GRAV=$(ls $STAGE_BUCKET/stardog-graviton*linux_amd64)

LAUNCH_NAME=$(cat $BUILD_DIR/name)
chmod 755 $THIS_DIR/$GRAV
export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$THIS_DIR/$BUILD_DIR

echo $LAUNCH_NAME

set +e
$THIS_DIR/$GRAV destroy --force $LAUNCH_NAME
if [ $? -ne 0 ]; then
    echo "FAILED"
    cat $STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR/deployments/$LAUNCH_NAME/logs/graviton.log
    exit 1
fi
exit 0
