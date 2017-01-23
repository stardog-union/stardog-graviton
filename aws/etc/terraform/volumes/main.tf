provider "aws" {
  region = "${var.aws_region}"
}

resource "aws_ebs_volume" "stardog_data" {
  availability_zone = "${element(var.aws_az[var.aws_region], count.index % length(var.aws_az[var.aws_region]))}"
  count = "${var.cluster_size}"
  size = "${var.storage_size}"
  tags {
    Name = "Stardog data volume"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}
