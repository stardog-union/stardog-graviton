#!/bin/bash

set -eu

SKIP=$1
export AWS_ACCESS_KEY_ID=$2
export AWS_SECRET_ACCESS_KEY=$3
STARDOG_VERSION=$4
GRAV_REPO=$5
LINUX_STAGE=$6
ENV_DIR=$7
OUT_DIR=$8

if [ $SKIP -eq 1 ]; then
    echo "Skipping the graviton tests"
    exit 0
fi

THIS_DIR=$(pwd)
GRAV_EXE=$(ls $THIS_DIR/$LINUX_STAGE/stardog-graviton-*)

echo $GRAV_EXE
chmod 755 $GRAV_EXE
export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$THIS_DIR/$ENV_DIR

LAUNCH_NAME=pipetest$RANDOM
echo $LAUNCH_NAME > $OUT_DIR/name

RELEASE_FILE=$THIS_DIR/$ENV_DIR/stardog-$STARDOG_VERSION.zip
LICENSE_FILE=$THIS_DIR/$ENV_DIR/stardog-license-key.bin

sed -e "s^@@LICENSE@@^$LICENSE_FILE^" -e "s^@@RELEASE@@^$RELEASE_FILE^" -e "s^@@VERSION@@^$STARDOG_VERSION^" $GRAV_REPO/ci/default.json.template > $ENV_DIR/default.json

ls -l $GRAV_EXE
set +e
$GRAV_EXE launch --force $LAUNCH_NAME
rc=$?
if [ $rc -ne 0 ]; then
    cat $STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR/deployments/$LAUNCH_NAME/logs/graviton.log
    $GRAV_EXE destroy --force $LAUNCH_NAME
    echo "FAILED"
    exit 1
fi

ls $STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR/
cp -r $STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR/* $OUT_DIR/
exit 0
