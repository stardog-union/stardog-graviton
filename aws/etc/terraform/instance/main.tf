provider "aws" {
  region = "${var.aws_region}"
}

provider "null" { }

data "aws_availability_zones" "available" {
}

resource "aws_vpc" "main" {
  cidr_block = "${var.internal_network}"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags {
    Name = "graviton instance VPC ${var.deployment_name}"
    Version = "${var.version}"
    DeploymentName = "${var.deployment_name}"
    StardogVirtualAppliance = "${var.deployment_name}"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "graviton gw ${var.deployment_name}"
    Version = "${var.version}"
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
