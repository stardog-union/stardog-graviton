import json
import os
import tempfile
import uuid
import subprocess
import sys
import time


def unzip_release(working_dir, sd_release):
    cmd = "unzip %s" % sd_release
    p = subprocess.Popen(cmd, shell=True, cwd=working_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to run cluster info")
    b = os.path.basename(sd_release)[:-4]
    return os.path.join(working_dir, b)


def get_status(exe_path, deployment_name):
    osf, f = tempfile.mkstemp(prefix="graviton_test_env", suffix=".json")
    p = subprocess.Popen("%s status %s --json-file %s" %
                         (exe_path, deployment_name, f),
                         shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to get status")
    print("Got the status of the deployment.")
    try:
        with os.fdopen(osf) as fptr:
            dict = json.load(fptr)
        print("Status dict %s" % dict)
        return str(dict['stardog_url'])
    finally:
        os.remove(f)


def run_basic_query_test(data_file, sd_dir, sd_url):
    sd_adim = os.path.join(sd_dir, "bin", "stardog-admin")
    cmd = "%s --server %s cluster info" % (sd_adim, sd_url)
    p = subprocess.Popen(cmd, shell=True)
    o, e = p.communicate()
    rc = p.wait()
    if rc != 0:
        print(cmd)
        print(e)
        print(o)
        raise Exception("Failed to run cluster info")
    db_name = "db%s" % str(uuid.uuid4()).split("-")[4]
    cmd = "%s --server %s db create --copy-server-side  -n %s %s" %\
          (sd_adim, sd_url, db_name, data_file)
    p = subprocess.Popen(cmd, shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to create the database")

    sdq = os.path.join(sd_dir, "bin", "stardog")
    cmd = "%s query %s/%s 'select ?s where {?s ?o ?p}'" % (sdq, sd_url, db_name)

    cnt = 0
    rc = 1
    while rc != 0:
        print("Try %d" % cnt)
        p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
        o, e = p.communicate()
        rc = p.wait()
        if rc != 0:
            print(e)
            print(o)
            if cnt > 2:
                raise Exception("Failed to run the query %s %d" % (cmd, rc))
            time.sleep(5)
        cnt = cnt + 1
    i = o.find('Query returned 462 results')
    if i < 0:
        raise Exception("Couldn't find 462 rows")


def main():
    configdir = sys.argv[1]
    release_full_path = sys.argv[2]
    graviton_exe = sys.argv[3]
    deployment_name = sys.argv[4]

    data_file = os.path.join(os.path.dirname(__file__), "rows.rdf")
    sd_dir = unzip_release(configdir, release_full_path)

    sd_url = get_status(graviton_exe, deployment_name)
    print("Stardog is running at %s" % sd_url)
    print("Start basic query tests")
    run_basic_query_test(data_file, sd_dir, sd_url)
    print("Start integration tests")
    return 0


if __name__ == "__main__":
    rc = main()
    sys.exit(rc)
