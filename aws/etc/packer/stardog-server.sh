#!/usr/bin/env bash

set -u

export STARDOG_HOME=/mnt/data/stardog-home
if [ -e /etc/stardog.env.sh ]; then
    source /etc/stardog.env.sh
fi

STARDOG_BIN=/usr/local/stardog/bin
PORT=5821
DAEMON="@@DAEMON@@"

start()
{
   echo "Starting stardog"
   ${STARDOG_BIN}/stardog-admin server start ${DAEMON} --home ${STARDOG_HOME} --port ${PORT}
   exit $?
}

stop()
{
   echo "Stopping stardog"
   ${STARDOG_BIN}/stardog-admin --server http://localhost:${PORT}/ server stop
   exit $?
}

case "$1" in
    start)
        start
        ;;

    stop)
        stop
        ;;

    restart|reload)
        stop
        start
        ;;

    *)
        echo "Usage: $0 {start|stop|restart}"
        exit 1;
esac
