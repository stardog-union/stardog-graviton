data "template_file" "stardog_zk_server" {
  count = "${var.zookeeper_size}"
  template = "$${host}:2181"
  vars {
    host = "${element(aws_elb.zookeeper.*.dns_name, count.index)}"
  }
}

data "template_file" "stardog_properties" {
  template = "${var.custom_stardog_properties}\n${file("stardog.properties.tpl")}"
  vars {
    zk_servers = "${join(",", data.template_file.stardog_zk_server.*.rendered)}"
    custom_data = "${var.custom_properties_data}"
  }
}

data "template_file" "stardog_userdata" {
  template = "${file("stardog_userdata.tpl")}"
  vars {
    custom_script = "${file(var.custom_script)}"
    stardog_conf = "${data.template_file.stardog_properties.rendered}"
    custom_log4j_data = "${var.custom_log4j_data}"
    deployment_name = "${var.deployment_name}"
    zk_servers = "${join(",", data.template_file.stardog_zk_server.*.rendered)}"
    environment_variables = "${var.environment_variables}"
    server_opts = "${var.stardog_start_opts}"
  }
}

resource "aws_autoscaling_group" "stardog" {
  count = "${var.stardog_size}"
  vpc_zone_identifier = ["${element(aws_subnet.stardog.*.id, count.index % length(aws_subnet.stardog.*.id))}"]
  name = "${var.deployment_name}sdasg${count.index}"
  max_size = "1"
  min_size = "1"
  desired_capacity = "1"
  launch_configuration = "${aws_launch_configuration.stardog.name}"
  load_balancers = ["${aws_elb.stardog.name}", "${aws_elb.stardoginternal.name}"]
  health_check_grace_period = 300
  health_check_type = "ELB"

  tag {
    key = "StardogVirtualAppliance"
    value = "${var.deployment_name}"
    propagate_at_launch = true
  }
  tag {
    key = "Name"
    value = "StardogNode"
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "stardog" {
  name_prefix = "${var.deployment_name}sdlc"
  image_id = "${var.baseami}"
  instance_type = "${var.stardog_instance_type}"
  user_data = "${data.template_file.stardog_userdata.rendered}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.stardog.id}"]
  iam_instance_profile = "${aws_iam_instance_profile.stardog.id}"
  associate_public_ip_address = true

  root_block_device {
    volume_type = "${var.root_volume_type}"
    volume_size = "${var.root_volume_size}"
    delete_on_termination = "true"
  }
}

resource "aws_iam_instance_profile" "stardog" {
  name = "${var.deployment_name}test_profile"
  role = "${aws_iam_role.stardog.name}"
}

resource "aws_iam_role" "stardog" {
  name = "${var.deployment_name}role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "stardog" {
  name = "${var.deployment_name}test_policy"
  role = "${aws_iam_role.stardog.id}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
              "ec2:Describe*",
              "ec2:Attach*"
            ],
            "Effect": "Allow",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "autoscaling:*",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
              "s3:List*",
              "s3:Get*"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}

resource "aws_security_group" "stardog" {
  name = "${var.deployment_name}sdsg"
  description = "Allow stardog traffic"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "stardog security group"
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

  # stardog port
  ingress {
    from_port = 5821
    to_port = 5821
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


resource "aws_security_group" "stardoglb" {
  name = "${var.deployment_name}sdlbsg"
  description = "Allow stardog traffic"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "stardog security group"
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

  # stardog port
  ingress {
    from_port = 5821
    to_port = 5821
    protocol = "tcp"
    cidr_blocks = ["${var.http_subnet}"]
  }

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "stardog" {
  count = "${length(data.aws_availability_zones.available.names)}"
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index + 100)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_elb" "stardog" {
  name = "${var.deployment_name}sdelb"
  subnets = ["${aws_subnet.stardog.*.id}"]
  security_groups = ["${aws_security_group.stardoglb.id}"]
  idle_timeout = "${var.elb_idle_timeout}"

  listener {
    instance_port     = 5821
    instance_protocol = "http"
    lb_port           = 5821
    lb_protocol       = "${var.external_protocol}"
    ssl_certificate_id = "${var.ssl_cert_arn}"
  }

  listener {
    instance_port     = 22
    instance_protocol = "tcp"
    lb_port           = 22
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 10
    timeout = 60
    target = "HTTP:5821/admin/healthcheck"
    interval = 300
  }

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_elb" "stardoginternal" {
  name = "${var.deployment_name}sdielb"
  subnets = ["${aws_subnet.stardog.*.id}"]
  security_groups = ["${aws_security_group.stardog.id}"]
  internal = true
  idle_timeout = "${var.elb_idle_timeout}"

  listener {
    instance_port     = 5821
    instance_protocol = "http"
    lb_port           = 5821
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 10
    timeout = 5
    target = "HTTP:5821/admin/healthcheck"
    interval = 15
  }

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}
