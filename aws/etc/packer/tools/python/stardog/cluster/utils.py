import json
import logging
import logging.config
import os
import psutil
import random
import requests
import subprocess
import time
import urllib
import urllib.request

import yaml

import boto3
import sys


def get_instance_ids(deply_name, count, region_name):
    found_instances = []
    client = boto3.client('autoscaling', region_name=region_name)
    for i in range(count):
        asg_name = "%ssdasg%d" % (deply_name, i)

        groups = client.describe_auto_scaling_groups(AutoScalingGroupNames=[asg_name])
        for g in groups['AutoScalingGroups']:
            for i in g['Instances']:
                found_instances.append(i['InstanceId'])

    return found_instances


def get_zk_instance_ids(deply_name, region_name):
    found_instances = []
    client = boto3.client('autoscaling', region_name=region_name)
    asg_name = "%szkasg" % deply_name

    groups = client.describe_auto_scaling_groups()
    for g in groups['AutoScalingGroups']:
        if g['AutoScalingGroupName'].find(asg_name) >= 0:
            for i in g['Instances']:
                found_instances.append(i['InstanceId'])

    return found_instances


def get_internal_ips_from_instance(instances, region_name):
    found_ips = []
    client = boto3.client('ec2', region_name=region_name)
    instances = client.describe_instances(InstanceIds=instances)
    for r in instances['Reservations']:
        for x in r['Instances']:
            found_ips.append(x['PrivateIpAddress'])
    return found_ips


def get_internal_ips_by_asg(deply_name, count, region_name):
    instance_ids = get_instance_ids(deply_name, count, region_name)
    return get_internal_ips_from_instance(instance_ids, region_name)


def get_internal_zk_ips_by_asg(deply_name, region_name):
    instance_ids = get_zk_instance_ids(deply_name, region_name)
    return get_internal_ips_from_instance(instance_ids, region_name)


def get_cluster_doc(sd_url, pw):
    full_url = sd_url + "/admin/cluster"
    logging.info("Trying to contact %s" % full_url)
    r = requests.get(sd_url + "/admin/cluster", auth=('admin', pw))
    if r.status_code != 200:
        raise Exception("Unable to get the cluster document %d" % r.status_code)
    return r.json()


def get_availability_zone():
    r = requests.get('http://169.254.169.254/latest/meta-data/placement/availability-zone')
    if r.status_code != 200:
        raise Exception("Unable to get the cluster document %d" % r.status_code)
    return r.text


def get_region():
    az = get_availability_zone()
    return az[:-1]


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


def find_volume(deployment_name, az=None):
    logging.debug("attempting to find an available volume for %s" % deployment_name)
    cmd = 'aws ec2 describe-volumes --filters Name=status,Values=available Name=tag-key,Values="DeploymentName" Name=tag-value,Values=%s' % deployment_name
    if az is not None:
        cmd = "%s Name=availability-zone,Values=%s" % (cmd, az)
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


def command(cmd, cmd_dir="./"):
    logging.debug("Running command: %s in directory %s" % (cmd, cmd_dir))
    p = subprocess.Popen(cmd, shell=True, cwd=cmd_dir)
    o, e = p.communicate()
    logging.debug("%s output: %s" % (cmd, o))
    logging.debug("%s error: %s" % (cmd, e))
    if p.returncode != 0:
        logging.warning("%s failed" % cmd)
        error = {'cmd': cmd, 'rc': p.returncode, 'output': o, 'error': e}
        return p.returncode, error
    return p.returncode, {}


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


def get_stardog_pid():
    for proc in psutil.process_iter():
        if proc.name() == 'java':
            for i in proc.cmdline():
                if i.find('stardog') >= 0:
                    return proc.pid
    raise Exception('Stardog process was not found')


def run_jstack(fname):
    pid = get_stardog_pid()
    cmd = "/usr/bin/jstack %d" % pid
    with open(fname, 'w') as fptr:
        p = subprocess.Popen(cmd, shell=True, stdout=fptr)
        o, e = p.communicate()
        rc = p.wait()
        if rc != 0:
            print(o)
            print(e)
            raise Exception('failed to get jstack')


def run_jmap(fname):
    pid = get_stardog_pid()
    cmd = "/usr/bin/jmap %d" % pid
    with open(fname, 'w') as fptr:
        p = subprocess.Popen(cmd, shell=True, stdout=fptr)
        o, e = p.communicate()
        rc = p.wait()
        if rc != 0:
            print(o)
            print(e)
            raise Exception('failed to get jstack')
