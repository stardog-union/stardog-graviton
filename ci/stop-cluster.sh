#!/bin/bash

set -eu

START_DIR=$(pwd)
OUTPUT_DIR=${START_DIR}/OUTPUT

GRAV_EXE=$(ls $OUTPUT_DIR/linux/stardog-graviton-*)

LAUNCH_NAME=$(cat $OUTPUT_DIR/name)
export STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR=$OUTPUT_DIR
echo $LAUNCH_NAME

ls -l $OUTPUT_DIR

$GRAV_EXE destroy --force $LAUNCH_NAME
