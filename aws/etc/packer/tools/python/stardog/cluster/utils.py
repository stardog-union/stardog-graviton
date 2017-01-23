import json
import logging
import logging.config
import os
import random
import subprocess
import time
import urllib
import urllib.request

import yaml


def get_meta_data(key):
    data = urllib.request.urlopen("http://169.254.169.254/latest/meta-data/%s" % key).read()
    return data.decode()


def wait_for_func(n, sleep_max, func):
    for i in range(0, n):
        if func():
            return True
        time.sleep(random.random() * sleep_max)
    return False


def run_json_command(cmd):
    p = subprocess.Popen(cmd, bufsize=0, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
    (sout, serr) = p.communicate()
    rc = p.wait()
    logging.debug("%s STDERR: %s" % (cmd, serr.decode()))
    logging.debug("%s STDOUT: %s" % (cmd, sout.decode()))
    if rc != 0:
        return False, None
    d = json.loads(sout.decode())
    return True, d


def find_volume(deployment_name):
    logging.debug("attempting to find an available volume for %s" % deployment_name)
    cmd = 'aws ec2 describe-volumes --filters Name=status,Values=available Name=tag-key,Values="DeploymentName" Name=tag-value,Values=%s' % deployment_name
    rc, vols = run_json_command(cmd)
    if not rc:
        logging.warning("A volume was not found for %s" % deployment_name)
        return None
    vols = vols["Volumes"]
    logging.debug("Volumes found %s" % vols)
    if len(vols) < 1:
        logging.warn("No volumes were found for %s" % deployment_name)
        return None
    logging.debug("Found the volume %s" % vols[0]["VolumeId"])
    return vols[0]["VolumeId"]


def volume_state(vol_id):
    cmd = 'aws ec2 describe-volumes --filters Name=volume-id,Values=%s' % vol_id
    rc, res = run_json_command(cmd)
    if not rc:
        return None
    try:
        vs = res["Volumes"][0]["State"]
        logging.debug("The volume state of %s %s" % (vol_id, vs))
        return vs
    except:
        logging.warn("Failed to get the volume state")
        return None


def attach_volume(volume_id, device, instance_id):
    logging.debug("Attempting to attach the volume %s to %s" % (volume_id, device))
    cmd = "aws ec2 attach-volume --volume-id %s --device %s --instance-id %s" % (volume_id, device, instance_id)
    rc, res = run_json_command(cmd)
    logging.debug("attach volume results %s %s" % (rc, res))
    return rc


def command(cmd):
    rc = subprocess.call(cmd, shell=True)
    return rc == 0


def setup_logging(logging_configfile=None):
    if logging_configfile is None:
        logging_configfile = "/usr/local/stardog-tools/python/logging.yaml"
    if not os.path.exists(logging_configfile):
        loghandler = logging.StreamHandler()
        top_logger = logging.getLogger("")
        top_logger.setLevel(logging.DEBUG)
        top_logger.addHandler(loghandler)
        return

    with open(logging_configfile, 'rt') as f:
        config = yaml.load(f.read())
        logging.config.dictConfig(config)
