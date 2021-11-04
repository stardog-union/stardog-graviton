#!/usr/bin/env bash

set -e

date > /tmp/boottime

export STARDOG_HOME=/mnt/data/stardog-home
${environment_variables}
echo '${environment_variables}' > /etc/stardog.env.sh
/usr/local/bin/stardog-find-volume ${deployment_name} /mnt/data /dev/xvdf

echo '${stardog_conf}' > $STARDOG_HOME/stardog.properties
MY_IP=`curl http://169.254.169.254/latest/meta-data/local-ipv4`
sed -i "s/@@LOCAL_IP@@/$MY_IP/" $STARDOG_HOME/stardog.properties

SERVER_OPTS="${server_opts}"

if [ -n $SERVER_OPTS ]; then
    sed -i "s/server start/server start $SERVER_OPTS/" /opt/stardog/stardog-server.sh
fi

/usr/local/bin/stardog-wait-for-socket 100 ${zk_servers}

set +e

if [ 'X${custom_log4j_data}' != 'X' ]; then
    echo '${custom_log4j_data}' | /usr/bin/base64 -d > $STARDOG_HOME/log4j2.xml
fi

systemctl enable stardog
systemctl start stardog
/usr/local/bin/stardog-wait-for-socket 60 localhost:5821

echo "Running the custom script..."
CUSTOM_SCRIPT=/tmp/custom
echo '${custom_script}' | /usr/bin/base64 -d > $CUSTOM_SCRIPT
chmod 755 $CUSTOM_SCRIPT
$CUSTOM_SCRIPT
echo "Done $?"

date >> /tmp/boottime
exit 0
