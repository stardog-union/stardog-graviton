import errno
import os
import logging
import tarfile
import tempfile
import subprocess
import sys

import stardog.cluster.utils as utils


def get_log(host, temp_dir, src_log, src_is_dir=False, node_type="stardog"):
    dst_dir = os.path.join(temp_dir, host, node_type)
    try:
        os.makedirs(dst_dir)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(dst_dir):
            pass
        else:
            raise

    new_dst = dst_dir
    if src_is_dir:
        name = os.path.basename(os.path.dirname(src_log))
        new_dst = os.path.join(dst_dir, name)
        try:
            os.makedirs(new_dst)
        except OSError:
            pass
    scp_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    cmd = "scp -r %s %s:%s %s" % (scp_opts, host, src_log, new_dst)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True, cwd=dst_dir)
    o, e = p.communicate()
    logging.info("Log copy stdout output %s" % o)
    logging.info("Log copy stderr output %s" % e)
    if p.returncode != 0:
        logging.warning("The log copy of %s %s failed" % (host, src_log))
        return False
    return True


def get_all_logs(ips, get_jstack=True, node_type="stardog", dst_dir=None):
    if dst_dir is None:
        dst_dir = tempfile.mkdtemp()
    logs_copied = 0
    for ip in ips:
        if get_jstack:
            try:
                ssh_cmd = "ssh %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sudo /usr/local/bin/stardog-jstack" % ip
                p = subprocess.Popen(ssh_cmd, shell=True)
                o, e = p.communicate()
                logging.info("jstack stdout output %s" % o)
                logging.info("jstack stderr output %s" % e)
                if p.returncode != 0:
                    logging.warning("jstack failed on %s" % ip)
            except Exception as ex:
                logging.error("Failed to get jstack info on %s" % ip, ex)

        b = get_log(ip, dst_dir, "/mnt/data/stardog-home/logs/*", src_is_dir=True, node_type=node_type)
        if b:
            logs_copied += 1
        source_log_files = [
            '/mnt/data/stardog-home/stardog.log*',
            '/zookeeper.log*',
            '/var/log/zookeeper.log*',
            '/var/log/syslog*',
            '/var/log/auth.log',
            '/var/log/kern.log',
            '/etc/stardog.env.sh',
            '/var/log/cloud-init.log',
            '/mnt/data/stardog-home/stardog.properties',
            '/var/log/cloud-init-output.log']
        for l in source_log_files:
            b = get_log(ip, dst_dir, l, node_type=node_type)
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
    n, log_dir = get_all_logs(ips, node_type="stardog")
    zk_ips = utils.get_internal_zk_ips_by_asg(deploy_name, region)
    m, log_dir = get_all_logs(zk_ips, get_jstack=False, node_type="zookeeper", dst_dir=log_dir)
    logging.info("Retrieved %d logs" % (n+m))
    if n < 1:
        raise Exception("No logs were gathered")
    create_tarball(log_dir, dst_file)
    return 0
