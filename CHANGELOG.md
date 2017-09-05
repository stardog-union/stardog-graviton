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
