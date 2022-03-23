#!/bin/bash -v

set -e

STARDOG_VERSION=${1}
IMAGE_VERSION=${2}

echo "Getting dependencies"
sudo apt-get update
sudo apt-get install python3-psutil unzip openjdk-8-jdk awscli python3 python3-pip jq libyaml-dev python3-yaml runit python3-boto3 -y

# Make some directories writable for the main user
sudo chmod 775 -R /usr/local
sudo chown -R ubuntu /usr/local

set +e
sudo mkdir /mnt/.tmp
sudo chown -R ubuntu /mnt/.tmp
sudo chmod 775 -R /mnt/.tmp
set -e

echo "Install stardog"
pushd /usr/local
	# todo - do not hardcode the address?
	mv /tmp/stardog.zip .
	unzip stardog.zip
	sd_dir=$(ls | `which grep` stardog-)
	if [ -z "${sd_dir}" ]; then
		echo "error setting up stardog"
		exit 1
	fi
    ls -ld /usr/local/
    ls -l /usr/local/
    ls -l /usr/local/${sd_dir}

	mv /usr/local/${sd_dir} /usr/local/stardog
popd

# make sure that stardog scripts have execution permissions
chmod 775 /usr/local/stardog/bin/stardog
chmod 775 /usr/local/stardog/bin/stardog-admin

# Create symlink to stardog scripts
sudo ln -s /usr/local/stardog/bin/stardog /usr/local/bin/stardog
sudo ln -s /usr/local/stardog/bin/stardog-admin /usr/local/bin/stardog-admin

# static http server for copying the stardog binaries
stardog_dir=/opt/stardog
sudo mkdir -p ${stardog_dir}
sudo chown -R ubuntu ${stardog_dir}
cd ${stardog_dir}
sudo chown -R ubuntu /usr/local/bin

echo "Installing Zookeeper"
# Install ZooKeeper
zookeeper_version=3.6.3
sudo wget --no-check-certificate https://dlcdn.apache.org/zookeeper/zookeeper-${zookeeper_version}/apache-zookeeper-${zookeeper_version}-bin.tar.gz && ls -alh && sudo tar -xvzf apache-zookeeper-${zookeeper_version}-bin.tar.gz
sudo mv apache-zookeeper-${zookeeper_version}-bin/ /usr/local/zookeeper-${zookeeper_version}
sudo chgrp -R ubuntu /usr/local/zookeeper-${zookeeper_version}
sudo chmod 775 -R /usr/local/zookeeper-${zookeeper_version}

# Create zk data directory
sudo mkdir -p /var/zkdata
sudo chgrp -R ubuntu /var/zkdata
sudo chmod 775 -R /var/zkdata

echo "Setting up version dependent information"
semver=( ${STARDOG_VERSION//./ } )
major="${semver[0]}"
minor="${semver[1]}"

if [ $major -gt 5 ]; then
    sudo sed -i 's/@@DAEMON@@/--daemon/' /tmp/stardog-server.sh
elif [[ $major -eq 5 && $minor -gt 1 ]]; then
    sudo sed -i 's/@@DAEMON@@/--daemon/' /tmp/stardog-server.sh
else
    sudo sed -i 's/@@DAEMON@@//' /tmp/stardog-server.sh
fi

echo "Success!"
