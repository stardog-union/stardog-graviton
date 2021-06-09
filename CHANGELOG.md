# 1.0.11 (2021-06-09)

Modifications:
* Fixed python tools install
* Updated terraform version

# 1.0.10 (2019-08-23)

Modifications:
* Removed ASGs and ELBs for ZooKeeper instances
* Gather additional log files for Stardog 7

Fixes:
* Preserve arguments passed to stardog server start
* Download ZooKeeper from archive instead of mirrors

# 1.0.9 (2018-06-05)

Fixes:
* Fixes bug that broke log gathering. Adds log gathering release test.

# 1.0.8 (2018-05-22)

Features:
* Added support for a volume from a snapshot id mounted on the bastion.
* Refactored the code layout to be more inline with Go best practices.

Fixes:
* Fixed a bug related to the linux cross compile.  When cross compile
  for Linux on OSX it results in the user.Current() function to not
  be available.  Now the release process will not cross compile for linux.

# 1.0.7 (2018-05-07)

Features:
* Added `update-stardog` command.  This updates a running cluster with a new
  version of Stardog making the development cycle much shorter.
* Allow JMX metrics reporting in stardog.properties.tpl.
* Updates terraform and packer to the latest versions.
* Fixes ZK log gathering, also gathers extra syslogs if it was rotated
* Configure bastion root volume with tf vars. Fixes (#54)

Fixes:
* Respect the --config-dir option.  Fixes (#67)
* Support long-form AMI. Fixes (#63)
* Updates ZK version since 3.4.11 was pulled from mirrors. Fixes (#61)

# 1.0.6 (2018-03-12)

Features:
* Use io1 volumes as the default instead of gp2.
* Gather all rotated out logs as well as current logs.
* Allow advanced setting of IO options with TF_VAR environment variables.

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
