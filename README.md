# Stardog Graviton: Virtual Appliance Manager

This application creates and controls a stardog database cluster running in
a cloud.  The current implementation only works in AWS.

Features:
- Create base images with all software dependencies burnt into them.
- Create a set of volumes to back a Stardog cluster. 
- Launch a fully working Stardog cluster database.
- Monitor health of the cluster.
- Cleanup all resources.

Dependencies:
- An [AWS](https://aws.amazon.com/) account. 
- [terraform](https://www.terraform.io/downloads.html) 0.7.9.
    - [Linux 64](https://releases.hashicorp.com/terraform/0.7.9/terraform_0.7.9_linux_amd64.zip)
    - [OSX](https://releases.hashicorp.com/terraform/0.7.9/terraform_0.7.9_darwin_amd64.zip)
- [packer](https://www.packer.io/downloads.html) 0.10.2.
    - [Linux](https://releases.hashicorp.com/packer/0.10.2/packer_0.10.2_linux_amd64.zip)
    - [OSX](https://releases.hashicorp.com/packer/0.10.2/packer_0.10.2_darwin_amd64.zip)
- A Stardog license.
- A Stardog release zip file (4.2.2 or later).
- The `stardog-graviton` single file executable.

## Quick start

In order to use `stardog-graviton` in its current form the following environment variables must be set.
```
AWS_ACCESS_KEY_ID=<a valid aws access key>
AWS_SECRET_ACCESS_KEY=<a valid aws secret key>
```

Both `terraform` and `packer` must be in your system path.

### Launch a cluster:

The easiest way to launch a cluster is to run `stardog-graviton` in interactive mode.  This will cause the program to ask a series of questions in order to get the needed values to launch a cluster.  Here is a sample session:

```
$ ./bin/stardog-graviton launch mystardog2
Failed to load the default file /Users/stardog/.graviton/default.json The file /Users/stardog/.graviton/default.json does not exist.
What version of stardog are you launching?: 4.2
What is the path to the Stardog release?: /Users/stardog/stardog-4.2.zip
There is no base image for version 4.2.
Do you wish to build one? (yes/no): yes
| Running packer to build the image...
done
AMI Successfully built: ami-0a1d486a
Creating the new deployment mystardog2
EC2 keyname (default): stardog
Private key path: /Users/stardog/.ssh/stardog
What is the path to your Stardog license?: /Users/stardog/data/stardog/stardog-license-key.bin
| Calling out to terraform to create the volumes...
- Calling out to terraform to stop builder instances...
Successfully created the volumes.
\ Creating the instance VMs...
Successfully created the instance.
Waiting for stardog to come up...
/ Waiting for external health check to pass...
The instance is healthy
Stardog is available here: http://mystardog2sdelb-682913646.us-west-1.elb.amazonaws.com:5821
ssh is available here: mystardog2belb-1558182568.us-west-1.elb.amazonaws.com
The instance is healthy
Coordinator:
   10.0.100.6:5821
Nodes:
   10.0.101.168:5821
   10.0.100.243:5821
Success.
```

To avoid being asked questions a file name ~/.graviton/default.json can be created.  An example can be found int the [defaults.json.example](defaults.json.example) file.

All of the components needed to run a Stardog cluster are considered part of a deployment.  Every deployment must be given a name that is unique to each cloud account.  In the above example the deployment name is `mystardog2`.

#### Status
Once the image has been successfully launched its health can be monitored with the `status` command:

```
$ stardog-graviton status mystardog2
Stardog is available here: http://mystardog2sdelb-1562381657.us-west-1.elb.amazonaws.com:5821
ssh is available here: mystardog2belb-1941074804.us-west-1.elb.amazonaws.com
The instance is healthy
Coordinator:
   10.0.100.110:5821
Nodes:
   10.0.101.230:5821
   10.0.100.33:5821
Success.
```

#### Cleanup
The EC2 charges by the hour for the VMs that Graviton runs thus when the cluster is no longer in use it is important to clean it up with the `destroy` commmand.

```
$ ./bin/stardog-graviton destroy mystardog2
Failed to load the default file /Users/stardog/.graviton/default.json The file /Users/stardog/.graviton/default.json does not exist.
This will destroy all volumes and instances associated with this deployment.
Do you really want to destroy? (yes/no): yes
- Deleting the instance VMs...
Successfully destroyed the instance.
\ Calling out to terraform to delete the images...
Successfully destroyed the volumes.
Success.
```

### Base image
The first time this is done a base image needs to be created.  This image will have Stardog, Zookeeper, and a set of other dependencies needed to run the cluster baked into it.  Even tho this image will be localized to your aws account no secrets (including the license) will be baked into it.  Future launches of the cluster with the same stardog version will not require this step.

To create the base image in a separate step use the following subcommands
```
  baseami [<flags>] <release> <sd-version>
    Create a base ami.
```


### Volumes
Every cluster needs a backing set of volumes to store the data.  In AWS this data is stored on [elastic block store](https://aws.amazon.com/ebs/) (ebs) volumes.  The launching a new cluster these volumes are created, formated and populated with your stardog license.  The database admin password is set at this time as well.  Because these volumes will contain your data and stardog licenses it is important to keep them secret.

To control the volumes use the following subcommands:

```
  volume new [<flags>] <deployment> <license> <size> <count>
    Create new backing storage.

  volume destroy [<flags>] <deployment>
    This will destroy the volumes permanently.

  volume status <deployment>
    Display information about the volumes.
```


### Instances
Running the stardog cluster requires several virtual machines.  At least 3 zookeeper nodes are needed for it to run safely and at least 2 stardog nodes.  Additionally a *bastion* node is used in order to allow ssh access to all other VMs as well as provide a configured client environment read to use.  AWS charges by the hour for the VMs so it is important to not leave them running.  In a given deployment the VMs can be started and stopped without destroying the data backing them.  The following subcommands can be used to control the VM instances:

```
  instance new [<flags>] <deployment> <zk>
    Create new set of VMs running Stardog.

  instance destroy [<flags>] <deployment>
    Destroy the instance.

  instance status [<flags>] <deployment>
    Get information about the instance.
```

### Cluster status
The status of a give deployment can be checked with the `status` subcommand.  The status can also be written to a json file if the --json-file option is included.  Here is an example session:
```
$ ./bin/stardog-graviton status mystardog2 --json-file=output.json
Failed to load the default file /Users/stardog/.graviton/default.json: The file /Users/stardog/.graviton/default.json does not exist.
Stardog is available here: http://mystardog2sdelb-682913646.us-west-1.elb.amazonaws.com:5821
ssh is available here: mystardog2belb-1558182568.us-west-1.elb.amazonaws.com
The instance is healthy
Coordinator:
   10.0.100.6:5821
Nodes:
   10.0.100.243:5821
   10.0.101.168:5821
Success.
```

The output file looks like the following:
```
{
    "stardog_url": "http://mystardog2sdelb-682913646.us-west-1.elb.amazonaws.com:5821",
    "ssh_host": "mystardog2belb-1558182568.us-west-1.elb.amazonaws.com",
    "healthy": true,
    "volume": {
        "VolumeIds": [
            "vol-c5070c6b",
            "vol-007183bf",
            "vol-2e070c80"
        ]
    },
    "instane": {
        "ZkNodesContact": [
            "internal-mystardog2zkelb0-770674234.us-west-1.elb.amazonaws.com",
            "internal-mystardog2zkelb1-1838829827.us-west-1.elb.amazonaws.com",
            "internal-mystardog2zkelb2-2057497106.us-west-1.elb.amazonaws.com"
        ]
    }
}
```

## Troubleshooting

### Logging

The `stardog-graviton` program logs to the console and to a log file.  To increase the level of console logging and `--verbose` to the command line multiple times and --log-level=DEBUG.  While this will provide details much more verbose logging can be found in the log file.  Each deployment will have its own log file located at ` ~/.graviton/deployments/<deployment name>/logs/graviton.log`

 ssh access to the cluster is provided via the bastion node.  Its contact point is displayed in the status command.  Once logged into that node the stardog nodes and zookeeper nodes can be access.  The following log files can be helpful in debugging a deployment that is not working:
 
 - /var/log/cloud-init-output.log
 - /var/log/stardog.image_config.log
 - /mnt/data/stardog-home/stardog.log
 - /mnt/data/stardog-home/zookeeper.log
 - /var/lib/cloud/instance/scripts/part-001
 - /var/log/cloud-init.log

### SSH Agent

Some of Graviton's features require a running [ssh-agent](https://en.wikipedia.org/wiki/Ssh-agent) loaded with the correct private key.

```
$ eval $(ssh-agent)
Agent pid 77228
$ ssh-add ~/.ssh/stardogbuzztroll
Identity added: /Users/stardog/.ssh/stardogbuzztroll (/Users/stardog/.ssh/stardogbuzztroll)
```

# Build stardog-graviton

go version 1.7.1 is required and must be in your system path in order to build `stardog-graviton` as is the program `make`.  Make sure that GOPATH is set properly,
and that graviton is checkout into $GOPATH/src/github.com/stardog-union

```
$ make
./scripts/build-local.sh
$ ls -l $GOPATH/bin/stardog-graviton
-rwxr-xr-x  1 stardog  staff  17812772 Nov  9 11:06 bin/stardog-graviton
```

# AWS architecture

This section describes the architecture of the Graviton when running in AWS.  Other cloud types may be added in the future.

## AWS Components
The following AWS components are the most important to the Graviton deployment.

### [Autoscaling Groups](https://aws.amazon.com/autoscaling/)
This is used to make sure that the zookeeper and stardog clusters are held to *N* nodes.  If a node is detected to be unhealthy AWS will restart that node

### [Elastic Load Balancer](https://aws.amazon.com/elasticloadbalancing/)
These are used for two reasons:
1. To distribute client requests across Stardog server nodes.
2. As a layer of indirection to make each node in an autoscaling group reliably addressable.  This is basically cruft due to some missing features in ec2.

### [Elastic Block Store](https://aws.amazon.com/ebs/)
This is used as a backing store for each Stardog node

## Interactions
Other AWS components are used as well but those are the key concepts.

The first things that is done for a deployment is create the volume set to back the database.  These are created and taged with the deployment name.  They remain in the users AWS account until they are deleted.  They do not have to be actively attached to a running VM.

When a deployment is launched the following autoscaling groups are created:
- One autoscaling group for *each* zookeeper node.  Each one of these autoscaling groups monitors a single node.  As mentioned above this is essentially a work around of some limited EC2 features.
- One autoscaling group for all the Stardog nodes.  This group makes sure N Stardog nodes are always running and healthy.  If one fails that VM will be restarted.
- One to make sure that a single bastion node is always running.

### Boot steps
1. First the zookeeper cluster comes up and each node joins a pool.
2. The Stardog nodes start.
    - Each Stardog node searches the EC2 account for an EBS volume tagged with the same deployment name.
    - Each node selects one of the found EBS volumes at random and attempts to mount it.
    - If it fails to mount (most likely due to one of the other nodes beating it to the mount) it retries *n* times.
    - Once it has a volume mounted to it verifies that it can find the zookeeper pool.
    - Finally it starts the start dog server.
3. The `stardog-graviton` program waits until the stardog cluser is healthy.  It does this by checking the url `http://<stardog load balancer>:5821/admin/healthcheck`
