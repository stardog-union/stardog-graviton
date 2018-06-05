import os
import shutil

import subprocess

import sys
import tempfile
import threading


_g_failed = []


def this_location():
    return os.path.abspath(os.path.dirname(__file__))


def checkenv(sd_license, release, ssh_key_path):
    required_vars = ['AWS_ACCESS_KEY_ID',
                     'AWS_SECRET_ACCESS_KEY',
                     'GOPATH']
    for k in required_vars:
        v = os.getenv(k)
        if v is None:
            raise Exception("The environment variable %s must be set" % k)
    p = subprocess.Popen("docker ps", shell=True)
    rc = p.wait()
    if rc != 0:
        raise Exception("The docker environment is not configured")

    file_list = [sd_license, release, ssh_key_path]
    for f in file_list:
        if not os.path.exists(f):
            raise Exception("The file %s does not exist" % f)
    os.unsetenv('STARDOG_ADMIN_PASSWORD')


def build_with_gox():
    base_dir = os.path.dirname(this_location())
    cmd = 'gox -osarch="linux/amd64" -osarch="darwin/amd64" ' \
          '-output=release/{{.OS}}_{{.Arch}}/stardog-graviton '\
          'github.com/stardog-union/stardog-graviton/cmd/stardog-graviton'
    p = subprocess.Popen(cmd, shell=True, cwd=base_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to cross compile graviton")
    if not os.path.exists(os.path.join(this_location(), "linux_amd64", "stardog-graviton")):
        raise Exception("The linux compile failed")
    if not os.path.exists(os.path.join(this_location(), "darwin_amd64",
                                       "stardog-graviton")):
        raise Exception("The osx compile failed")


def prep_run(sd_license, release, grav_exe, ssh_key_path):
    src_dir = this_location()
    work_dir = tempfile.mkdtemp(prefix="graviton",
                                dir=os.path.abspath(os.path.dirname(__file__)))

    try:
        files_to_join_and_copy = ['rows.rdf', 'smoke_test_1.py']
        for f in files_to_join_and_copy:
            shutil.copy(os.path.join(src_dir, f),
                        os.path.join(work_dir, f))

        shutil.copy(sd_license,
                    os.path.join(work_dir, "stardog-license-key.bin"))
        shutil.copy(release,
                    os.path.join(work_dir, os.path.basename(release)))
        shutil.copy(grav_exe,
                    os.path.join(work_dir, "stardog-graviton"))
        shutil.copy(ssh_key_path,
                    os.path.join(work_dir, "ssh_key"))
        return work_dir
    finally:
        pass


def run_local(work_dir, ssh_key_name, release):
    print("Running in %s" % work_dir)
    cmd = "python %s %s %s %s %s" % (
        os.path.join(work_dir, "smoke_test_1.py"),
        work_dir, release, ssh_key_name, os.path.dirname(this_location()))
    print("Running %s" % cmd)
    p = subprocess.Popen(cmd, shell=True, cwd=work_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to run the smoke test")
    print ("XXX Local run was successful")


def build_docker(image_name):
    print("Building the docker container")
    cmd = "docker build -t %s . --no-cache" % image_name
    p = subprocess.Popen(cmd, shell=True, cwd=this_location())
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed build the container")


def compile_linux(image_name):
    print("Compiling in a docker container")
    top_dir = os.path.join(this_location(), "..")
    try:
        os.makedirs(os.path.join(this_location(), "release", "linux_amd64"))
    except:
        pass

    internal_gopath = "/opt/go/src/"
    docker_cmd = "/usr/lib/go-1.10/bin/go build -o %s/src/github.com/stardog-union/stardog-graviton/release/linux_amd64/stardog-graviton github.com/stardog-union/stardog-graviton/cmd/stardog-graviton" % internal_gopath
    cmd = "docker run -e GOPATH=%s -v %s:%s/src/github.com/stardog-union/stardog-graviton -it %s %s" % (internal_gopath, top_dir, internal_gopath, image_name, docker_cmd)
    print(cmd)
    p = subprocess.Popen(cmd, shell=True, cwd=this_location())
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed build the container")


def run_docker(work_dir, ssh_key_name, release, image_name):
    print("Running docker for testing...")
    cmd = "docker run -v %s:/smoke " \
          "-e AWS_SECRET_ACCESS_KEY=%s " \
          "-e AWS_ACCESS_KEY_ID=%s " \
          "-it %s " \
          "python /smoke/smoke_test_1.py /smoke %s %s" %\
          (work_dir,
           os.environ['AWS_SECRET_ACCESS_KEY'],
           os.environ['AWS_ACCESS_KEY_ID'],
           image_name, release, ssh_key_name)
    p = subprocess.Popen(cmd, shell=True, cwd=work_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to run the smoke tests in the container")


def print_usage():
    print("Invalid arguments:")
    print("<path to stardog license> <path to stardog release file>"
          " <path to ssh private key> <aws key name>")


def get_version():
    cmd = "git describe --abbrev=0 --tags"
    work_dir = os.path.dirname(this_location())
    p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, cwd=work_dir)
    (o, e) = p.communicate()
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to zip the file")
    return o.strip()


def zip_one(arch):
    ver = get_version()
    work_dir = os.path.join(this_location(), arch)
    cmd = "zip stardog-graviton_%s_%s.zip stardog-graviton" % (ver, arch)
    p = subprocess.Popen(cmd, shell=True, cwd=work_dir)
    rc = p.wait()
    if rc != 0:
        raise Exception("Failed to zip the file")


def darwin_test(sd_license, release, ssh_key_path, ssh_key_name):
    try:
        darwin_binary = os.path.join(this_location(),
                                      "darwin_amd64", "stardog-graviton")
        release_name = os.path.basename(release)
        work_dir = prep_run(sd_license, release, darwin_binary, ssh_key_path)
        run_local(work_dir, ssh_key_name, release_name)
        print("Successfully smoke tested for darwin.")
        print("Exe: darwin_amd64/stardog-graviton")
    except Exception as ex:
        global _g_failed
        _g_failed.append("Darwin failed: %s" % str(ex))
        print("TEST ERROR darwin %s" % str(ex))
    zip_one("darwin_amd64")


def linux_test(sd_license, release, ssh_key_path, ssh_key_name):
    try:
        build_docker("graviton-release-tester")
        compile_linux("graviton-release-tester")
        linux_binary = os.path.join(this_location(),
                                    "linux_amd64", "stardog-graviton")
        release_name = os.path.basename(release)
        work_dir = prep_run(sd_license, release, linux_binary, ssh_key_path)
        run_docker(work_dir, ssh_key_name, release_name, "graviton-release-tester")
        print("Successfully smoke tested for darwin.")
        print("Exe: linux_amd64/stardog-graviton")
    except Exception as ex:
        global _g_failed
        _g_failed.append("Linus failed: %s" % str(ex))
        print("TEST ERROR linux %s" % str(ex))
    zip_one("linux_amd64")


def main():
    if len(sys.argv) < 4:
        print_usage()
        return 1
    sd_license = sys.argv[1]
    release = sys.argv[2]
    ssh_key_path = sys.argv[3]
    ssh_key_name = sys.argv[4]

    checkenv(sd_license, release, ssh_key_path)
    build_with_gox()
    threads = []
    if sys.platform != "darwin":
        print("XXXXXX We cannot test of OSX on this platform")
    else:
        t = threading.Thread(
                target=darwin_test,
                args=(sd_license, release, ssh_key_path, ssh_key_name))
        threads.append(t)
        t.start()

    t = threading.Thread(
            target=linux_test,
            args=(sd_license, release, ssh_key_path, ssh_key_name))
    threads.append(t)
    t.start()

    print("Started %d tests, waiting for completion..." % len(threads))
    for t in threads:
        t.join()

    if len(_g_failed) != 0:
        print("The tests failed %s" % _g_failed)
        return 1
    print("Success!")
    return 0


if __name__ == "__main__":
    rc = main()
    sys.exit(rc)
