 #!/usr/bin/env bash

set -e

DEVICE=$1
MOUNT_POINT=$2

mkfs -t ext4 /dev/xvdh
mkdir -p /mnt/data
mount /dev/xvdh /mnt/data
mkdir -p /mnt/data/stardog-home
chown -R ubuntu /mnt/data