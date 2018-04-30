import datetime
import logging
import os
import subprocess
import sys

import stardog.cluster.utils as utils


# Stardog should be stopped before refreshing the binaries,
# see update_stardog.py for the general process.
def main():
    cur_time = datetime.datetime.now().strftime('%Y%m%d-%H%M%S')

    release_file = sys.argv[1]
    logging.debug("Release file: %s" % release_file)
    base_zip_file = os.path.basename(release_file)
    logging.debug("Base zip file: %s" % base_zip_file)
    base_file = base_zip_file.rstrip('.zip')
    logging.debug("Base file: %s" % base_file)

    errors = []
    cp_cmd = "cp %s /usr/local/" % release_file
    rc, err = utils.command(cp_cmd, cmd_dir="/usr/local/")
    if rc != 0:
        errors.append(err)

    unzip_cmd = "unzip /usr/local/%s" % base_zip_file
    rc, err = utils.command(unzip_cmd, cmd_dir="/usr/local/")
    if rc != 0:
        errors.append(err)

    backup_cmd = "mv /usr/local/stardog /usr/local/stardog.%s" % cur_time
    rc, err = utils.command(backup_cmd, cmd_dir="/usr/local/")
    if rc != 0:
        errors.append(err)

    mv_cmd = "mv /usr/local/%s /usr/local/stardog" % base_file
    rc, err = utils.command(mv_cmd, cmd_dir="/usr/local/")
    if rc != 0:
        errors.append(err)

    if errors:
        raise Exception(errors)

    return 0
