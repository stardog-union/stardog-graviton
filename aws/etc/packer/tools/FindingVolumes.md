# Finding Volumes

This document describes how values will be found for the Stardog virtual appliance running in AWS.

## Overview

All Stardog cluster nodes are run in AWS auto-scaling group.  Auto-scaling groups make sure that N VMs are always running.  If a VM is determined to have failed it is restarted.  Because of this VMs cannot be statically provisioned with volumes.  They must have a way to attach and mount a volume when they start up.

In order to accomplish this a `startup script` is associated with each VM.  Whenever a VM is booted this script is run via [cloud-init](https://help.ubuntu.com/community/CloudInit) .  One of its primary responsibilities is finding a volume to mount for use as a backing store for the Stardog instance.

## Volumes

Every virtual appliance is pre-provisioned with a set of volumes.  There is a one to one mapping of volumes to stardog nodes but they are loosely coupled, ie: each Stardog node must have exactly one volume but which specific volume it has will not matter.  If a customer wishes to increase or decrease the size of their cluster they must expand or reduce the size of the volume pool as well.

The AWS Elastic Block Store ( [EBS](https://aws.amazon.com/ebs/) ) is used to create the volumes.  Every virtual appliance instance will have a distinct `deployment name` associated with it.  Each EBS volume used with a virtual appliance is tagged with this deployment name.  When a Stardog VM starts up it searches for exactly on of these volumes and attempts to attach it and mount it.  If the volume is not yet formatted it will format it and set it up for use with Stardog.
 
 ## Find Volume script
 
 Race conditions come up when many VMs boot at once and search for volumes to attach but thankfully AWS will only allow 1 VM to attach a volume at a time.  The find volume script first gets a list of all volumes with the `deployment name` tag.  It then picks one and attempts to attach it.  If it fails if waits a random amount of time and then repeats the process.
 
 ## AWS IAM
 
 In order for the Stardog VM to search for and attach volumes it must have access to AWS credentials.  As part of the terraform deployment each VM is given newly created IAM credentials with just enough access to find and attach volumes.  As we move forward we should find ways to lock this down even further.