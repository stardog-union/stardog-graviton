
resource "aws_security_group" "stardog_data" {
  name = "${var.deployment_name}sddsg"
  description = "Allow stardog traffic"

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

resource "aws_instance" "stardog_data" {
  availability_zone = "${element(var.aws_az[var.aws_region], count.index % length(var.aws_az[var.aws_region]))}"
  count = "${var.cluster_size}"
  tags {
    Name = "Volume builder"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
  instance_type = "${var.instance_type}"
  ami = "${var.ami}"
  key_name = "${var.aws_key_name}"
  security_groups = ["${aws_security_group.stardog_data.name}"]
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
      "sudo mkdir -p /mnt/data/stardog-home",
      "sudo chown -R ubuntu /mnt/data/"
    ]
  }

  provisioner "file" {
    source = "${var.stardog_license}"
    destination = "/mnt/data/stardog-home/stardog-license-key.bin"
  }

  depends_on = ["aws_volume_attachment.stardog_data"]
}

