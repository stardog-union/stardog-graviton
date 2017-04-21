variable "storage_size" {
  type = "string"
  description = "The size of the volume in gigabytes."
}

variable "cluster_size" {
  type = "string"
  description = "The number of stardog nodes to use (must be odd and greater than 1)."
}

variable "aws_region" {
  type = "string"
  description = "The AWS region to create things in."
}

variable "aws_az" {
  type = "map"
  description = "The AWS availability zone."
  default = {
    us-east-1 = [
      "us-east-1a",
      "us-east-1b",
      "us-east-1c",
      "us-east-1d",
      "us-east-1e"]
    us-east-2 = [
      "us-east-2a",
      "us-east-2b",
      "us-east-2c"]
    us-west-1 = [
      "us-west-1a",
      "us-west-1c"]
    us-west-2 = [
      "us-west-2a",
      "us-west-2b",
      "us-west-2c"],
    eu-central-1 = [
      "eu-central-1a",
      "eu-central-1b"
    ],
    eu-west-1 = [
      "eu-west-1a",
      "eu-west-1b",
      "eu-west-1c"
    ]
  }
}

variable "deployment_name" {
  type = "string"
  description = "A string that is unique to this stardog data in a given account and resource"
}

variable "aws_key_name" {
  type = "string"
  description = "The AWS key name"
}

variable "key_path" {
  type = "string"
  description = "The path to the private key"
}

variable "ami" {
  type = "string"
  description = "The ami to use for building the image"
}

variable "instance_type" {
  type = "string"
  description = "The instance type for formating the volumes"
}

variable "stardog_license" {
  type = "string"
  description = "The path to your stardog license"
}

variable "volume_type" {
  type = "string"
  description = "The EBS storage type"
  default = "io1"
}

variable "iops" {
  type = "string"
  description = "The IOPS to provision the volume with"
}
