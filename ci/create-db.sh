#!/bin/bash

set -eu

START_DIR=$(pwd)
OUTPUT_DIR=${START_DIR}/OUTPUT

GRAV_EXE=$(ls $OUTPUT_DIR/linux/stardog-graviton-*)

LAUNCH_NAME=$(cat $OUTPUT_DIR/name)
RELEASE_FILE=$OUTPUT_DIR/stardog-$STARDOG_VERSION.zip

export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$OUTPUT_DIR
LAUNCH_NAME=$(cat $OUTPUT_DIR/name)

python ./ci/create_db.py $OUTPUT_DIR $RELEASE_FILE $GRAV_EXE $LAUNCH_NAME
if [ $? -ne 0]; then
    echo "Fail"
    exit 1
fi
echo "Success"
exit 0
