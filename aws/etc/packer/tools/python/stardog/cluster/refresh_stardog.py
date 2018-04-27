import datetime
import logging
import os
import subprocess
import sys


def run_cmd(cmd, dir="/usr/local/"):
    print(cmd)
    p = subprocess.Popen(cmd, shell=True, cwd=dir)
    o, e = p.communicate()
    logging.info("%s output: %s" % (cmd, o))
    logging.info("%s error: %s" % (cmd, e))
    if p.returncode != 0:
        logging.warning("%s failed" % cmd)
        error = {'cmd': cmd, 'rc': p.returncode, 'output': o, 'error': e}
        return p.returncode, error
    return p.returncode, {}


def main():
    cur_time = datetime.datetime.now().strftime('%Y%m%d-%H%M%S')

    release_file = sys.argv[1]
    print(release_file)

    base_zip_file = os.path.basename(release_file)
    print(base_zip_file)
    base_file = base_zip_file.rstrip('.zip')
    print(base_file)

    errors = []
    rc, err = run_cmd("sudo systemctl stop stardog")
    if rc != 0:
        errors.append(err)

    rc, err = run_cmd("mv /usr/local/stardog /usr/local/stardog.%s" % cur_time)
    if rc != 0:
        errors.append(err)

    rc, err = run_cmd("cp %s /usr/local/" % release_file)
    if rc != 0:
        errors.append(err)

    rc, err = run_cmd("unzip /usr/local/%s" % base_zip_file)
    if rc != 0:
        errors.append(err)

    rc, err = run_cmd("mv /usr/local/%s /usr/local/stardog" % base_file)
    if rc != 0:
        errors.append(err)

    rc, err = run_cmd("sudo systemctl start stardog")
    if rc != 0:
        errors.append(err)

    if errors:
        raise Exception(errors)

    return 0
