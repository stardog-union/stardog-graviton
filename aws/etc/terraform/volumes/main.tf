provider "aws" {
  region = "${var.aws_region}"
}

data "aws_availability_zones" "available" {
}


resource "aws_ebs_volume" "stardog_data" {
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index % length(data.aws_availability_zones.available.names))}"
  count = "${var.cluster_size}"
  size = "${var.stardog_home_volume_size}"
  type = "${var.stardog_home_volume_type}"
  iops = "${var.stardog_home_volume_type == "io1" ? var.stardog_home_volume_iops : 0}"
  tags {
    Name = "Stardog data volume"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}
