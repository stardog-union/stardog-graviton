import stardog.cluster.utils as utils


def main():
    errors = []
    rc, err = utils.command("sudo systemctl start stardog", cmd_dir="/usr/local/")
    if rc != 0:
        errors.append(err)

    if errors:
        raise Exception(errors)

    return 0
