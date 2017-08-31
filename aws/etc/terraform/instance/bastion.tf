
resource "aws_autoscaling_group" "bastion" {
  vpc_zone_identifier = ["${aws_subnet.bastion.*.id}"]
  name = "${var.deployment_name}basg"
  max_size = "1"
  min_size = "1"
  desired_capacity = "1"
  launch_configuration = "${aws_launch_configuration.bastion.name}"
  load_balancers       = ["${aws_elb.bastion.name}"]
  health_check_grace_period = 90
  health_check_type = "ELB"

  tag {
    key = "StardogVirtualAppliance"
    value = "${var.deployment_name}"
    propagate_at_launch = true
  }
  tag {
    key = "Name"
    value = "BastionNode"
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "bastion" {
  name_prefix = "${var.deployment_name}blc"
  image_id = "${var.baseami}"
  instance_type = "${var.bastion_instance_type}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.bastion.id}"]
  associate_public_ip_address = true
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

resource "aws_elb" "bastion" {
  name = "${var.deployment_name}belb"

  # The same availability zone as our instances
  subnets = ["${aws_subnet.bastion.*.id}"]
  security_groups = ["${aws_security_group.bastion.id}"]

  listener {
    instance_port     = 22
    instance_protocol = "tcp"
    lb_port           = 22
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    timeout = 3
    target = "TCP:22"
    interval = 30
  }

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_subnet" "bastion" {
  count = "${length(var.aws_az[var.aws_region])}"
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index + 200)}"
  availability_zone = "${element(var.aws_az[var.aws_region], count.index)}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}