#!/bin/bash

set -eu

SKIP=$1
export AWS_ACCESS_KEY_ID=$2
export AWS_SECRET_ACCESS_KEY=$3
STARDOG_VERSION=$4
GRAV_REPO=$5
LINUX_STAGE=$6
ENV_DIR=$7
LAUNCH_OUTPUT=$8

if [ $SKIP -eq 1 ]; then
    echo "Skipping the graviton tests"
    exit 0
fi

THIS_DIR=$(pwd)
GRAV_EXE=$(ls $THIS_DIR/$LINUX_STAGE/stardog-graviton-*)
LAUNCH_NAME=$(cat $LAUNCH_OUTPUT/name)
chmod 755 $GRAV_EXE

RELEASE=$THIS_DIR/$ENV_DIR/stardog-$STARDOG_VERSION.zip
export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$THIS_DIR/$LAUNCH_OUTPUT
LAUNCH_NAME=$(cat $LAUNCH_OUTPUT/name)

export STARDOG_HOME=$THIS_DIR/$ENV_DIR
python $GRAV_REPO/ci/create_db.py $THIS_DIR/$ENV_DIR $RELEASE $GRAV_EXE $LAUNCH_NAME $GRAV_REPO
if [ $? -ne 0]; then
    echo "Fail"
    exit 1
fi
echo "Success"
exit 0
