import datetime
import logging
import socket
import sys

import http.server

import stardog.cluster.utils as utils


g_activation_time = datetime.datetime.now() + datetime.timedelta(minutes=10)


class BaseHandler(http.server.BaseHTTPRequestHandler):
    def tester(self):
        return False

    def do_GET(self):
        logging.info("Received a get request from %s", self.address_string())

        if self.tester():
            self.send_response(200)
            self.send_header('Content-type','text/html')
            self.end_headers()
            self.wfile.write("OK".encode(encoding="ascii"))
            logging.info("The service is healthy")
        else:
            self.send_response(500)
            self.send_header('Content-type','text/html')
            self.end_headers()
            self.wfile.write("FAILED".encode(encoding="ascii"))
            logging.info("The service is NOT healthy")
        return


class ZookeeperHandler(BaseHandler):
    zookeeper_host = "localhost"
    zookeeper_port = 2181
    zookeeper_message = "ZooKeeper instance is not currently serving requests"

    def tester(self):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.connect((self.zookeeper_host, self.zookeeper_port))
            s.sendall('stat\n'.encode(encoding='ascii'))
            data = s.recv(5*1024)
            message = data.decode(encoding="ascii")
            logging.debug("from zk %s" % message)
            rc = message.find(self.zookeeper_message)
            s.close()
            n = datetime.datetime.now()
            # Have a softer health check at the beginning so that routing will happen
            if n < g_activation_time:
                return True
            return rc < 0
        except Exception as ex:
            logging.warning("Error: %s" % str(ex))
            return False


def main():
    utils.setup_logging()

    port = int(sys.argv[1])
    try:
        logging.info("Listening for HTTP monitoring on %d", port)
        server = http.server.HTTPServer(('', port), ZookeeperHandler)
        server.serve_forever()
    except KeyboardInterrupt:
        print ('^C received, shutting down the web server')
        server.socket.close()