resource "aws_instance" "bastion" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

	timeouts {
	  create = "30m"
	  delete = "30m"
	  update = "30m"
	}

	associate_public_ip_address = true
	instance_type = "${var.bastion_instance_type}"
	ami = "${var.baseami}"
	key_name = "${var.aws_key_name}"
	security_groups = ["${aws_security_group.bastion.id}"]
	iam_instance_profile = "${aws_iam_instance_profile.stardog.id}"
	subnet_id = "${aws_subnet.bastion.id}"

  root_block_device {
    volume_type = "${var.bastion_root_volume_type}"
    volume_size = "${var.bastion_root_volume_size}"
    iops = "${var.bastion_root_volume_type == "io1" ? var.bastion_root_volume_iops : 0}"
    delete_on_termination = "true"
  }

  tags {
	  Name = "BastionNode"
	  DeploymentName = "${var.deployment_name}"
	  StardogVirtualAppliance = "${var.deployment_name}"
	}
}

resource "aws_security_group" "bastion" {
  name = "${var.deployment_name}bsg"
  description = "Allow ssh traffic"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "bastion single security group"
    Version = "${var.version}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }

  # allow ssh from anywhere
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "bastion" {
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index + 200)}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}