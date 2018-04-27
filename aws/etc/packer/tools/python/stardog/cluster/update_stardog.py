import logging
import subprocess
import sys

import stardog.cluster.utils as utils


def upload_file(ip, upload_file):
    scp_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    cmd = "scp -r %s %s %s:%s" % (scp_opts, upload_file, ip, upload_file)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True)
    o, e = p.communicate()
    logging.info("Upload file stdout %s" % o)
    logging.info("Upload file stderr %s" % e)
    if p.returncode != 0:
        logging.warning("The upload of %s %s failed" % (ip, upload_file))
        error = {'cmd': cmd, 'rc': p.returncode, 'output': o, 'error': e}
        return p.returncode, error
    return p.returncode, {}


def refresh_stardog(ip, release_file):
    ssh_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    refresh_cmd = "/usr/local/bin/stardog-refresh"
    cmd = "ssh %s %s '%s %s'" % (ssh_opts, ip, refresh_cmd, release_file)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True)
    o, e = p.communicate()
    logging.info("stardog-refresh stdout %s" % o)
    logging.info("stardog-refresh stderr %s" % e)
    if p.returncode != 0:
        logging.warning("Stardog refresh on %s failed" % ip)
        error = {'cmd': cmd, 'rc': p.returncode, 'output': o, 'error': e}
        return p.returncode, error
    return p.returncode, {}


def main():
    deploy_name = sys.argv[1]
    print(deploy_name)
    count = int(sys.argv[2])
    print(count)
    release_file = sys.argv[3]
    print(release_file)

    region = utils.get_region()
    ips = utils.get_internal_ips_by_asg(deploy_name, count, region)

    errors = []
    for ip in ips:
        print(ip)
        rc, err = upload_file(ip, release_file)
        if rc != 0:
            errors.append(err)
        rc, err = refresh_stardog(ip, release_file)
        if rc != 0:
            errors.append(err)

    if errors:
        raise Exception(errors)

    return 0
