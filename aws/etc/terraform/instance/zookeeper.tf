
data "template_file" "zk_server" {
  count = "${var.zookeeper_size}"
  template = "${file("server-spec.tpl")}"
  vars {
    id = "${count.index + 1}"
    host = "${element(aws_elb.zookeeper.*.dns_name, count.index)}"
  }
}

data "template_file" "zk_conf" {
  template = "${file("zoo.cfg.tpl")}"
  vars {
    zk_server_list = "${join("", data.template_file.zk_server.*.rendered)}"
  }
}

data "template_file" "zk_health_ping" {
  count = "${var.zookeeper_size}"
  template = "/usr/local/bin/stardog-wait-for-pgm 100 ping -c 1 $${host}"
  vars {
    id = "${count.index + 1}"
    host = "${element(aws_elb.zookeeper.*.dns_name, count.index)}"
  }
}

data "template_file" "zk_userdata" {
  count = "${var.zookeeper_size}"
  template = "${file("zk_userdata.tpl")}"
  vars {
    custom_zk_script = "${file(var.custom_zk_script)}"
    index = "${count.index + 1}"
    zk_conf = "${data.template_file.zk_conf.rendered}"
    zk_health_wait = "${join("\n", data.template_file.zk_health_ping.*.rendered)}"
  }
}

resource "aws_autoscaling_group" "zookeeper" {
  name = "${var.deployment_name}zkasg${count.index}"
  count = "${var.zookeeper_size}"
  vpc_zone_identifier = ["${element(aws_subnet.zk.*.id, count.index % length(aws_subnet.zk.*.id))}"]
  max_size = 1
  min_size = 1
  desired_capacity = 1
  launch_configuration = "${element(aws_launch_configuration.zookeeper.*.name, count.index)}"
  health_check_grace_period = "${var.zk_health_grace_period}"
  health_check_type = "ELB"
  load_balancers = ["${element(aws_elb.zookeeper.*.name, count.index)}"]

  tag {
    key = "StardogVirtualAppliance"
    value = "${var.deployment_name}"
    propagate_at_launch = true
  }
    tag {
    key = "Name"
    value = "ZookeeperNode"
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "zookeeper" {
  count = "${var.zookeeper_size}"
  name_prefix = "${var.deployment_name}zklc${count.index}"
  image_id = "${var.baseami}"
  instance_type = "${var.zk_instance_type}"
  user_data = "${element(data.template_file.zk_userdata.*.rendered, count.index)}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.zookeeper.id}"]
  # XXX TODO figure out why we need a public ip for external routing
  associate_public_ip_address = true
}

resource "aws_elb" "zookeeper" {
  name = "${var.deployment_name}zkelb${count.index}"

  count = "${var.zookeeper_size}"
  internal = true

  subnets = ["${element(aws_subnet.zk.*.id, count.index % length(aws_subnet.zk.*.id))}"]
  security_groups = ["${aws_security_group.zookeeper.id}"]

  listener {
    instance_port     = 2181
    instance_protocol = "tcp"
    lb_port           = 2181
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 2888
    instance_protocol = "tcp"
    lb_port           = 2888
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 3888
    instance_protocol = "tcp"
    lb_port           = 3888
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 9000
    instance_protocol = "tcp"
    lb_port           = 9000
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 22
    instance_protocol = "tcp"
    lb_port           = 22
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold = "${var.zk_healthy_threshold}"
    unhealthy_threshold = "${var.zk_unhealthy_threshold}"
    timeout = "${var.zk_health_timeout}"
    target = "HTTP:9000/"
    interval = "${var.zk_health_interval}"
  }

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_security_group" "zookeeper" {
  name = "${var.deployment_name}zksg"
  description = "Allow stardog traffic"
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
  count = "${length(data.aws_availability_zones.available.names)}"
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}
