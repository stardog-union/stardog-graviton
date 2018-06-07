
data "template_file" "zk_server" {
  count = "${var.zookeeper_size}"
  template = "${file("server-spec.tpl")}"
  vars {
    id = "${count.index + 1}"
    host = "${format("10.0.%d.%d", count.index + 200, count.index + 200)}"
  }
}

data "template_file" "zk_conf" {
  template = "${file("zoo.cfg.tpl")}"
  vars {
    zk_tickTime="${var.zk_tickTime}"
    zk_initLimit="${var.zk_initLimit}"
    zk_syncLimit="${var.zk_syncLimit}"
    zk_dataDir="${var.zk_dataDir}"
    zk_dataLogDir="${var.zk_dataLogDir}"
    zk_server_list = "${join("", data.template_file.zk_server.*.rendered)}"
  }
}

data "template_file" "zk_health_ping" {
  count = "${var.zookeeper_size}"
  template = "/usr/local/bin/stardog-wait-for-pgm 100 ping -c 1 $${host}"
  vars {
    id = "${count.index + 1}"
    host = "${format("10.0.%d.%d", count.index + 200, count.index + 200)}"
  }
}

data "template_file" "zk_userdata" {
  count = "${var.zookeeper_size}"
  template = "${file("zk_userdata.tpl")}"
  vars {
    environment_variables = "${var.environment_variables}"
    custom_zk_script = "${file(var.custom_zk_script)}"
    index = "${count.index + 1}"
    zk_conf = "${data.template_file.zk_conf.rendered}"
    zk_health_wait = "${join("\n", data.template_file.zk_health_ping.*.rendered)}"
  }
}

resource "aws_instance" "zookeeper" {
  count = "${var.zookeeper_size}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  private_ip = "${format("10.0.%d.%d", count.index + 200, count.index + 200)}"

  timeouts {
	  create = "30m"
	  delete = "30m"
	  update = "30m"
	}

  ami = "${var.baseami}"
  instance_type = "${var.zk_instance_type}"
  user_data = "${element(data.template_file.zk_userdata.*.rendered, count.index)}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.zookeeper.id}"]
  iam_instance_profile = "${aws_iam_instance_profile.stardog.id}"
  subnet_id = "${element(aws_subnet.zk.*.id, count.index)}"

  root_block_device {
    volume_type = "${var.zk_root_volume_type}"
    volume_size = "${var.zk_root_volume_size}"
    iops = "${var.zk_root_volume_type == "io1" ? var.zk_root_volume_iops : 0}"
    delete_on_termination = "true"
  }

  tags {
	  Name = "ZookeeperNode"
	  DeploymentName = "${var.deployment_name}"
	  StardogVirtualAppliance = "${var.deployment_name}"
	}
}

resource "aws_security_group" "zookeeper" {
  name = "${var.deployment_name}zksg"
  description = "Allow zookeeper traffic"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "stardog zookeeper security group"
    Version = "${var.version}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }

  # allow ssh from anywhere
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["${var.internal_network}"]
  }

  # allow pings
  ingress {
    from_port = 8
    to_port = 0
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # zookeeper ports
  ingress {
    from_port = 2181
    to_port = 2181
    protocol = "tcp"
    cidr_blocks = ["${var.internal_network}"]
  }

  ingress {
    from_port = 2888
    to_port = 2888
    protocol = "tcp"
    cidr_blocks = ["${var.internal_network}"]
  }

  ingress {
    from_port = 9000
    to_port = 9000
    protocol = "tcp"
    cidr_blocks = ["${var.internal_network}"]
  }

  ingress {
    from_port = 3888
    to_port = 3888
    protocol = "tcp"
    cidr_blocks = ["${var.internal_network}"]
  }

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "zk" {
  count = "${var.zookeeper_size}"
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index + 200)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}
