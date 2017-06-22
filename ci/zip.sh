#!/bin/bash

set -u

START_DIR=$(pwd)
OUTPUT_DIR=${START_DIR}/OUTPUT

THIS_DIR=$(pwd)
LINUX_GRAV_EXE=$(ls $OUTPUT_DIR/linux/stardog-graviton-*)
DARWIN_GRAV_EXE=$(ls $OUTPUT_DIR/darwin/stardog-graviton-*)
DESTNAME=$(basename $LINUX_GRAV_EXE)

if [ "X$TAG_VERSION" != "X" ]; then
    VER=$TAG_VERSION
else
    VER=$(cat etc/version)
fi

echo $VER
$LINUX_GRAV_EXE --version
$LINUX_GRAV_EXE --version  2>&1 | grep $VER
if [ $? -ne 0 ]; then
    echo "The version information is not correct in the artifact"
    exit 1
fi

echo "make the dirs"
mkdir -p $OUTPUT_DIR/$VER

echo "Create the linux zip"
cp $LINUX_GRAV_EXE stardog-graviton
chmod 755 stardog-graviton
zip "$OUTPUT_DIR/$VER/stardog-graviton_$VER""_linux_amd64.zip" stardog-graviton
echo "Create the darwin zip"
cp $DARWIN_GRAV_EXE stardog-graviton
chmod 755 stardog-graviton
zip "$OUTPUT_DIR/$VER/stardog-graviton_$VER""_darwin_amd64.zip" stardog-graviton
ls $OUTPUT_DIR/$VER
