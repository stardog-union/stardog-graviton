#!/bin/bash

set -eu

START_DIR=$(pwd)
OUTPUT_DIR=${START_DIR}/OUTPUT

STARDOG_RELEASE_URL=https://complexible.artifactoryonline.com/complexible/stardog-binaries/complexible/stardog/stardog-${STARDOG_VERSION}.zip

RELEASE_FILE=$OUTPUT_DIR/stardog-$STARDOG_VERSION.zip
LICENSE_FILE=$OUTPUT_DIR/stardog-license-key.bin

echo ${STARDOG_RELEASE_URL}
curl -u ${artifactoryUsername}:${artifactoryPassword} ${STARDOG_RELEASE_URL} > $RELEASE_FILE

sed -e "s^@@LICENSE@@^$LICENSE_FILE^" -e "s^@@RELEASE@@^$RELEASE_FILE^" -e "s^@@VERSION@@^$STARDOG_VERSION^" ./ci/default.json.template > $OUTPUT_DIR/default.json

echo $STARDOG_LICENSE | base64 -d > $LICENSE_FILE

if [ "X$AMI" != "X" ]; then
    echo '{"us-west-2":"'$AMI'"}' > $OUTPUT_DIR/amis-${STARDOG_VERSION}.json
fi
