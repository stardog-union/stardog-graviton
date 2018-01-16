import datetime

import stardog.cluster.utils as utils


def main():
    fname = "/mnt/data/stardog-home/logs/jstack.%s.log" % datetime.datetime.now().strftime('%Y-%m-%d_%m')
    utils.run_jstack(fname)
    fname = "/mnt/data/stardog-home/logs/jmap.%s.log" % datetime.datetime.now().strftime('%Y-%m-%d_%m')
    utils.run_jmap(fname)
    return 0
