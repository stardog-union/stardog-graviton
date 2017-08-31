import json
import os
import subprocess
import threading

_g_failed = []
_g_lock = threading.Lock()

_run_count = 2


def run_region(region):
    try:
        deploy_name = "bz%d%s" % (_run_count, region.replace("-", ""))
        cmd = "~/go/bin/stardog-graviton launch --force --region %s %s" % (region, deploy_name)
        p = subprocess.Popen(cmd, shell=True)
        out, err = p.communicate()
        prc = p.wait()
        if prc != 0:
            with _g_lock:
                _g_failed.append(region + " failed to start " + deploy_name)
            print("%s failed %d" % (region, prc))
            return

        cmd = "~/go/bin/stardog-graviton destroy --force %s" % deploy_name
        p = subprocess.Popen(cmd, shell=True, stderr=subprocess.PIPE)
        out, err = p.communicate()
        prc = p.wait()
        if prc != 0:
            with _g_lock:
                _g_failed.append(region + " NOT DESTROYED!")
            print("%s failed" % region)
            print(err)
    except Exception as ex:
        with _g_lock:
            _g_failed.append(region + "WITH EXCEPTION")



with open(os.path.expanduser("~/.graviton/base-amis-5.0.2.json")) as fptr:
    d = json.load(fptr)

zones = ["ap-northeast-1", "ap-southeast-2"]
thread_list = []
for r in d.keys():
    print(r)
    t = threading.Thread(target=run_region, args=(r,))
    thread_list.append(t)
    t.start()

for t in thread_list:
    t.join()

for f in _g_failed:
    print("%s failed" % f)
