# 1.0.5 (2018-01-16)

Features:
* Add jmap output to the Stardog node log fetching.
* Use the --daemon flag soun Stardog version 5.2.0 and higher to avoid systemd
  restarts while nodes are joining.
* Set environment variables on Zookeeper nodes as well as Stardog nodes.

# 1.0.4 (2018-01-02)

Features:
* Make internal LB host for Stardog easily accessible on stdout #43
* End logs with .log instead of ip address of node #42
* Get ZK logs as part of log gathering #41
* Validate deployment name against AWS naming rules #30
* Run Stardog nodes under systemd with auto-restart.
* Switched AWS Autoscaling health check to EC2 checks.  This helps prevent
  ASG from rebooting VMs during node joins.
* Added a version check for base AMIs.
* Added the ability to control Zookeeper's configuration.

# 1.0.3 (2017-11-27)

Bug fixes:
* Query AWS for availability zones.
* Add timeout to the creation of the instance builder VMs.

# 1.0.2 (2017-11-20)

Feature:
* Allow the fetching of logs from nodes that failed to join the cluster.
* Added the ability to run a custom script on stardog nodes.
* Upgraded the version of Zookeeper used to 3.4.11

Bug fixes:
* Make the version a proper semver string.

# 1.0.1 (2017-09-05)

Feature
* Updated the Ubuntu base AMIs to 16.04 HVM images.
* Added logic to download packer and terraform if they are not found.
* Added logic to check the versions of packer and terraform and to cache good ones when found.

# 1.0.0 (2017-06-22)

# Feature
* Added a means to control the Stardog log4j file.

# 0.9.0-beta2 (2017-04-28)

Bug fixes:
* Reported the wrong version number.

# 0.9.0-beta1 (2017-04-27)

Initial release
---------------

Features:
* Launch a fully working Stardog cluster into AWS with a single command.
* Automatically monitor health and repair failed nodes.
* Manage multiple Stardog cluster deployments.
* Pause running resources without destroying data.
