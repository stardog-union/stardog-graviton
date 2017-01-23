import logging
import os
import sys

import stardog.cluster.utils as utils


def find_mount_iteration(deployment_name, device, instance_id):
    volume_id = utils.find_volume(deployment_name)
    if volume_id is None:
        return False
    logging.debug("Attaching the volume %s..." % volume_id)
    rc = utils.attach_volume(volume_id, device, instance_id)
    if not rc:
        logging.warning("Failed to attach the volume %s..." % volume_id)
        return False
    logging.info("Attached the volume %s." % volume_id)

    # wait for the device to appear
    def wait_for_file():
        logging.info("Checking for the file %s" % device)
        return os.path.exists(device)
    rc = utils.wait_for_func(10, 10, wait_for_file)
    if not rc:
        return False

    # wait for AWS to say it is attached
    def wait_for_attached():
        logging.debug("Checking the volume state of %s" % volume_id)
        return utils.volume_state(volume_id) == "in-use"
    rc = utils.wait_for_func(3, 5, wait_for_attached)

    return rc


def mount_format(device, mount_point):
    try:
        os.makedirs(mount_point)
    except FileExistsError:
        pass
    logging.debug("Attempting to mount, may not be formatted yet")
    if not utils.command("mount %s %s" % (device, mount_point)):
        logging.info("Failed to mount the disk, formatting first")
        if not utils.command("mkfs -t ext4 %s" % device):
            e_msg = "Failed to format the disk"
            logging.error(e_msg)
            raise Exception(e_msg)
        if not utils.command("mount %s %s" % (device, mount_point)):
            e_msg = "Failed to mount the disk after format"
            logging.error(e_msg)
            raise Exception(e_msg)
        logging.info("Disk mounted, making dirs and setting permissions...")
        home_dir = os.path.join(mount_point, "stardog-home")
        try:
            os.makedirs(home_dir, 0o775)
        except FileExistsError:
            pass
        utils.command("chown -R ubuntu %s" % home_dir)


def main():
    utils.setup_logging()

    try:
        instance_id = utils.get_meta_data('instance-id')
        az = utils.get_meta_data('placement/availability-zone')

        region = az[:-1]
        os.environ["AWS_DEFAULT_REGION"] = region
        deployment_name = sys.argv[1]
        mount_point = sys.argv[2]
        device = sys.argv[3]

        def find_attach_it():
            return find_mount_iteration(deployment_name, device, instance_id)
        rc = utils.wait_for_func(10, 5, find_attach_it)
        if not rc:
            e_msg = "Failed to attach the volume in the given number of tries"
            logging.error(e_msg)
            raise Exception(e_msg)

        logging.info("Successfully attached the volume, now mounting...")
        mount_format(device, mount_point)
        logging.info("Success")
    except Exception as ex:
        logging.error("An error occured %s")
        raise

    return 0