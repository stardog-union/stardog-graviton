import os
import logging
import tarfile
import tempfile
import subprocess
import sys

import stardog.cluster.utils as utils


def get_log(host, temp_dir, src_log, src_is_dir=False):
    if src_is_dir:
        name = os.path.basename(os.path.dirname(src_log))
        dst_name = os.path.join(temp_dir, name + "." + host)
        try:
            os.makedirs(dst_name)
        except OSError:
            pass
    else:
        name = os.path.basename(src_log)
        dst_name = os.path.join(temp_dir, name + "." + host)
    scp_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    cmd = "scp -r %s %s:%s %s" % (scp_opts, host, src_log, dst_name)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True, cwd=temp_dir)
    o, e = p.communicate()
    logging.info("Log copy stdout output %s" % o)
    logging.info("Log copy stderr output %s" % e)
    if p.returncode != 0:
        logging.warning("The log copy of %s %s failed" % (host, src_log))
        return False
    return True


def get_all_logs(ips):
    dst_dir = tempfile.mkdtemp()
    logs_copied = 0
    for ip in ips:
        b = get_log(ip, dst_dir, "/mnt/data/stardog-home/stardog.log")
        if b:
            logs_copied += 1
        b = get_log(ip, dst_dir, "/mnt/data/stardog-home/logs/*", src_is_dir=True)
        if b:
            logs_copied += 1
        source_log_files = [
            '/var/log/syslog', 
            '/var/log/auth.log',
            '/var/log/kern.log',
            '/var/log/cloud-init.log',
            '/var/log/cloud-init-output.log']
        for l in source_log_files:
            b = get_log(ip, dst_dir, l)
            if b:
                logs_copied += 1
    return logs_copied, dst_dir


def create_tarball(src_dir, dst_file):
    with tarfile.open(dst_file, 'w:gz') as tar:
        tar.add(src_dir, arcname='stardog_logs')


def main():
    deploy_name = sys.argv[1]
    count = int(sys.argv[2])
    dst_file = sys.argv[3]

    region = utils.get_region()
    ips = utils.get_internal_ips_by_asg(deploy_name, count, region)
    n, log_dir = get_all_logs(ips)
    logging.info("Retrieved %d logs" % n)
    if n < 1:
        raise Exception("No logs were gathered")
    create_tarball(log_dir, dst_file)
    return 0
