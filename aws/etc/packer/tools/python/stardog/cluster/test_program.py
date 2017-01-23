import logging
import subprocess
import sys

import stardog.cluster.utils as utils


def run_program(cmd, tries):
    def pgm_func():
        try:
            p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            o, e = p.communicate()
            logging.debug("STDOUT: %s", o.decode())
            logging.debug("STDERR: %s", e.decode())
            rc = p.wait()
            if rc == 0:
                logging.info("The program %s succeeded", cmd)
                return True
            else:
                logging.warning("The program %s failed %d", cmd, rc)
        except Exception as ex:
            logging.warning("There was an exception running %s.  %s.", cmd, ex)
            return False
    logging.info("Start the program run loop for the command %s")
    return utils.wait_for_func(tries, 30, pgm_func)


def main():
    utils.setup_logging()
    tries = int(sys.argv[1])
    cmd = sys.argv[2:]
    rc = run_program(' '.join(cmd), tries)
    if not rc:
        return 1
    return 0