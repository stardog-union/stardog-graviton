#!/bin/bash

set -eu

SKIP=$1
export AWS_ACCESS_KEY_ID=$2
export AWS_SECRET_ACCESS_KEY=$3
GRAV_REPO=$4
LINUX_STAGE=$5
ENV_DIR=$6
LAUNCH_OUTPUT=$7

if [ $SKIP -eq 1 ]; then
    echo "Skipping the graviton tests"
    exit 0
fi

ls
echo "BUILD"

echo "OLD"

THIS_DIR=$(pwd)
GRAV_EXE=$(ls $THIS_DIR/$LINUX_STAGE/stardog-graviton-*)

LAUNCH_NAME=$(cat $LAUNCH_OUTPUT/name)
chmod 755 $GRAV_EXE
export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$THIS_DIR/$LAUNCH_OUTPUT

echo $LAUNCH_NAME

$GRAV_EXE destroy --force $LAUNCH_NAME
