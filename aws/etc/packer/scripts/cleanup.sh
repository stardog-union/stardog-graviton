#!/usr/bin/env bash

apt-get -y autoremove
apt-get -y clean

echo "cleaning up dhcp leases"
rm /var/lib/dhcp/*

echo "cleaning up udev rules"
rm -rf /etc/udev/rules.d/70-persistent-net.rules
rm -rf /dev/.udev/
rm -f /lib/udev/rules.d/75-persistent-net-generator.rules

set -e
echo "Cleaning up cloud-init data..."
rm -rf /var/lib/cloud/instances
rm -rf /var/lib/cloud/instance
rm -f /home/ubuntu/.ssh/authorized_keys
rm -f /var/log/cloud-init*
