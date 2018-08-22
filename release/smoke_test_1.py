import json
import os
import tempfile
import uuid
import subprocess
import sys


def start_sd(exe_path, extra_args, deployment_name):
    p = subprocess.Popen("%s launch --force %s %s" % (exe_path, extra_args, deployment_name),
                         shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to start stardog")
    return


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


def stop_sd(exe_path, deployment_name):
    p = subprocess.Popen("%s destroy %s --force" % (exe_path, deployment_name),
                         shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to destroy the deployment")


def run_basic_query_test(conf_path, sd_dir, sd_url):
    data_file = os.path.join(conf_path, "rows.rdf")
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
    p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
    o, e = p.communicate()
    rc = p.wait()
    if rc != 0:
        print(e)
        print(o)
        raise Exception("Failed to run the query %s %d" % (cmd, rc))
    i = o.find('Query returned 462 results')
    if i < 0:
        raise Exception("Couldn't find 462 rows")


def run_log_gather_test(graviton_exe, deployment_name):
    output_file = "/tmp/stardog-logs-%s.tar.gz" % deployment_name
    cmd = "%s logs %s --output-file=%s" % (graviton_exe, deployment_name, output_file)
    p = subprocess.Popen(cmd, shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to gather logs for the deployment")
    if not os.path.isfile(output_file):
        raise Exception("Failed to gather logs: output file does not exist")
    if os.stat(output_file).st_size <= 0:
        raise Exception("Failed to gather logs: output file is empty")
    print("Successfully gathered logs")


def run_security_test(sd_dir, sd_url, security_enabled=True):
    sd_adim = os.path.join(sd_dir, "bin", "stardog-admin")
    cmd = "%s --server %s server status -u baduser -p badpassword" % (sd_adim, sd_url)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True)
    o, e = p.communicate()
    rc = p.wait()
    print("return code=%s" % rc)
    print("output=%s" % o)
    print("error=%s" % e)
    # command with bad user/password should fail with security enabled
    if security_enabled and rc == 0:
        raise Exception("Stardog command with bad user/password succeeded with security enabled")
    # command with bad user/password should pass with security disabled
    if not security_enabled and rc != 0:
        raise Exception("Stardog command with bad user/password failed even though security is disabled")
    print("Security check passed (security_enabled=%s)" % security_enabled)


def make_defaults_file(working_dir, sd_license, release_full_path, release_name,
                       ssh_key_path, ssh_key_name):
    suffix = ".zip"
    if not release_name.endswith(suffix):
        raise Exception("The license file is not properly formatted "
                        "(should end with %s)" % suffix)

    token = "stardog-"
    x = release_name.find(token)
    if x < 0:
        raise Exception("The license file was not properly formatted "
                        "(It should contain the string '%s'" % token)
    version = release_name[x + len(token):-4]

    j = {
        "license_path": os.path.join(working_dir, sd_license),
        "private_key": ssh_key_path,
        "log_level": "DEBUG",
        "cloud_type": "aws",
        "volume_size": 4,
        "cluster_size": 3,
        "release_file": release_full_path,
        "zookeeper_size": 3,
        "sd_version": version,
        "http_mask": "0.0.0.0/0",
        "cloud_options": {
            "region": "us-west-1",
            "zk_instance_type": "t2.small",
            "sd_instance_type": "t2.medium",
            "aws_key_name": ssh_key_name
        }
    }
    with open(os.path.join(working_dir, "default.json"), mode="w") as fptr:
        json.dump(j, fptr)


def setup(conf_path):
    os.environ['STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR'] = conf_path


def unzip_release(working_dir, sd_release):
    cmd = "unzip %s" % sd_release
    p = subprocess.Popen(cmd, shell=True, cwd=working_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to run cluster info")
    b = os.path.basename(sd_release)[:-4]
    return os.path.join(working_dir, b)


def run_integration_tests(source_dir, sd_url):
    cmd = "make test"
    new_env = os.environ.copy()
    new_env['STARDOG_DESCRIPTION_PATH'] = sd_url
    p = subprocess.Popen(cmd, shell=True, env=new_env, cwd=source_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to run the integration tests")


def deploy_and_run_tests(working_dir, source_dir, sd_dir,
                         deployment_name, security_enabled=True):
    extra_args = ""
    if not security_enabled:
        extra_args = "--disable-security"
    graviton_exe = os.path.join(working_dir, "stardog-graviton")
    start_sd(graviton_exe, extra_args, deployment_name)
    try:
        print("Started the deployment %s" % deployment_name)
        sd_url = get_status(graviton_exe, deployment_name)
        print("Stardog is running at %s" % sd_url)
        print("Start security test")
        run_security_test(sd_dir, sd_url, security_enabled)
        print("Start basic query tests")
        run_basic_query_test(working_dir, sd_dir, sd_url)
        if os.environ.get('SSH_AUTH_SOCK'):
            print("Start log gather test")
            run_log_gather_test(graviton_exe, deployment_name)
        else:
            print("SSH AGENT IS NOT RUNNING, SKIPPING LOG GATHERING TEST")
        print("Start integration tests")
        if source_dir is not None:
            run_integration_tests(source_dir, sd_url)
    except Exception as ex:
        print("Failed: %s" % str(ex))
        raise
    finally:
        print("Cleaning up %s" % deployment_name)
        stop_sd(graviton_exe, deployment_name)


def main():
    working_dir = sys.argv[1]
    release = sys.argv[2]
    ssh_key_name = sys.argv[3]
    source_dir = None
    if len(sys.argv) > 4:
        source_dir = sys.argv[4]

    release_full_path = os.path.join(working_dir, release)
    sd_license = os.path.join(working_dir, "stardog-license-key.bin")
    ssh_key_path = os.path.join(working_dir, "ssh_key")

    make_defaults_file(working_dir, sd_license, release_full_path, release,
                       ssh_key_path, ssh_key_name)
    sd_dir = unzip_release(working_dir, release_full_path)

    setup(working_dir)
    deployment_name = "gravtest%s" % str(uuid.uuid4()).split("-")[4]
    deploy_and_run_tests(working_dir, source_dir, sd_dir, deployment_name, security_enabled=True)
    deploy_and_run_tests(working_dir, source_dir, sd_dir, deployment_name, security_enabled=False)

    return 0


if __name__ == "__main__":
    rc = main()
    sys.exit(rc)
