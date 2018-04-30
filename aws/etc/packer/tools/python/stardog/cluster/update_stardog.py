import logging
import subprocess
import sys

import stardog.cluster.utils as utils


def upload_file(ip, upload_file):
    scp_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    cmd = "scp -r %s %s %s:%s" % (scp_opts, upload_file, ip, upload_file)
    return utils.command(cmd)


def refresh_stardog_binaries(ip, release_file):
    ssh_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    refresh_cmd = "/usr/local/bin/stardog-refresh-binaries"
    cmd = "ssh %s %s '%s %s'" % (ssh_opts, ip, refresh_cmd, release_file)
    return utils.command(cmd)


def stop_stardog(ip):
    ssh_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    refresh_cmd = "/usr/local/bin/stardog-stop"
    cmd = "ssh %s %s '%s'" % (ssh_opts, ip, refresh_cmd)
    return utils.command(cmd)


def start_stardog(ip):
    ssh_opts = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    refresh_cmd = "/usr/local/bin/stardog-start"
    cmd = "ssh %s %s '%s'" % (ssh_opts, ip, refresh_cmd)
    return utils.command(cmd)


def main():
    deploy_name = sys.argv[1]
    logging.debug("Deployment name: %s" % deploy_name)
    count = int(sys.argv[2])
    logging.debug("Count: %s" % count)
    release_file = sys.argv[3]
    logging.debug("Release file: %s" % release_file)

    region = utils.get_region()
    ips = utils.get_internal_ips_by_asg(deploy_name, count, region)

    errors = []
    for ip in ips:
        rc, err = upload_file(ip, release_file)
        if rc != 0:
            errors.append(err)

    for ip in ips:
        rc, err = stop_stardog(ip)
        if rc != 0:
            errors.append(err)

    for ip in ips:
        rc, err = refresh_stardog_binaries(ip, release_file)
        if rc != 0:
            errors.append(err)

    for ip in ips:
        rc, err = start_stardog(ip)
        if rc != 0:
            errors.append(err)

    if errors:
        raise Exception(errors)

    return 0
