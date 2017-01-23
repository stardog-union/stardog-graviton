import logging
import socket
import sys

import stardog.cluster.utils as utils


def test_connection(host_port, tries):
    ha = host_port.split(":")
    host = ha[0]
    port = int(ha[1])

    def ping_func():
        logging.debug("Trying to connect to %s" % host_port)
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(1.0)
        try:
            s.connect((host, port))
        except Exception as ex:
            logging.warning("Socket error %s:%d | %s" % (host, port, ex))
            return False
        logging.info("Successfully connected to %s:%d" % (host, port))
        return True

    logging.info("Start connection tries to %s" % host_port)
    return utils.wait_for_func(tries, 30, ping_func)


def main():
    utils.setup_logging()

    tries = int(sys.argv[1])
    host_port_list_str = sys.argv[2]

    # put this in many threads later
    for hp in host_port_list_str.split(","):
        rc = test_connection(hp, tries)
        if not rc:
            raise Exception("Failed to form the connection")
            logging.info("Made a connection to %s" % hp)
    return 0