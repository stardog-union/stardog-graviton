resource "aws_volume_attachment" "bastion_volume" {
	device_name = "/dev/xvdh"
	volume_id = "${aws_ebs_volume.bastion_volume.id}"
	instance_id = "${aws_instance.bastion.id}"
	force_detach = "true"
	depends_on = ["aws_instance.bastion", "aws_ebs_volume.bastion_volume"]
}

resource "aws_ebs_volume" "bastion_volume" {
	availability_zone = "${aws_subnet.bastion.availability_zone}"
	snapshot_id = "%s"
	type = "${var.bastion_volume_type}"
	iops = "${var.bastion_volume_type == "io1" ? var.bastion_volume_iops : 0}"
	tags {
	  Name = "Bastion volume"
	  DeploymentName = "${var.deployment_name}"
	  StardogVirtualAppliance = "${var.deployment_name}"
	}
	depends_on = ["aws_instance.bastion"]
}

resource "null_resource" "bastion_volume_provisioner" {
	triggers {
	  bastion_instance_id = "${aws_instance.bastion.id}"
	}

	connection {
	  user = "ubuntu"
	  host = "${aws_instance.bastion.public_dns}"
	  private_key = "${file("%s")}"
	  agent = false
	  timeout = "10m"
	}

	provisioner "remote-exec" {
	  inline = [
		"set -e",
		"sudo mkdir -p /mnt/data",
		"sudo mount /dev/xvdh /mnt/data",
		"sudo chown -R ubuntu /mnt/data/"
	  ]
	}

	depends_on = ["aws_volume_attachment.bastion_volume"]
}