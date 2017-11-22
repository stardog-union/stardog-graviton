
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags {
    Name = "graviton volume builder ${var.deployment_name}"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "stardog virtual appliance gw ${var.deployment_name}"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
  route_table_id         = "${aws_vpc.main.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.gw.id}"
}

resource "aws_security_group" "stardog_data" {
  name = "${var.deployment_name}sddsg"
  description = "Allow stardog traffic"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "stardog data security group"
    StardogVirtualAppliance = "${var.deployment_name}"
  }

  # allow ssh from anywhere
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "stardog" {
  count = "${var.cluster_size}"
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "${format("10.0.%d.0/24", count.index)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_instance" "stardog_data" {
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index % length(data.aws_availability_zones.available.names))}"
  count = "${var.cluster_size}"
  tags {
    Name = "Volume builder"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
  associate_public_ip_address = true
  instance_type = "${var.instance_type}"
  ami = "${var.ami}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.stardog_data.id}"]
  subnet_id = "${element(aws_subnet.stardog.*.id, count.index)}"
  depends_on = ["aws_ebs_volume.stardog_data"]
}

resource "aws_volume_attachment" "stardog_data" {
  count = "${var.cluster_size}"
  device_name = "/dev/xvdh"
  volume_id = "${element(aws_ebs_volume.stardog_data.*.id, count.index)}"
  instance_id = "${element(aws_instance.stardog_data.*.id, count.index)}"
}

resource "null_resource" "stardog_data" {
  count = "${var.cluster_size}"

  # Settings for SSH connection
  connection {
    user = "ubuntu"
    host = "${element(aws_instance.stardog_data.*.public_dns, count.index)}"
    private_key = "${file(var.key_path)}"
    agent = false
    timeout = "10m"
  }

  provisioner "remote-exec" {
    inline = [
      "set -e",
      "sudo mkfs -t ext4 /dev/xvdh",
      "sudo mkdir -p /mnt/data",
      "sudo mount /dev/xvdh /mnt/data",
      "sudo mkdir -p /mnt/data/stardog-home/logs",
      "sudo chown -R ubuntu /mnt/data/"
    ]
  }

  provisioner "file" {
    source = "${var.stardog_license}"
    destination = "/mnt/data/stardog-home/stardog-license-key.bin"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo umount /mnt/data/"
    ]
  }

  depends_on = ["aws_volume_attachment.stardog_data"]
}

