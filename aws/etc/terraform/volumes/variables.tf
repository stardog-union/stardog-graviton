variable "cluster_size" {
  type = "string"
  description = "The number of stardog nodes to use (must be odd and greater than 1)."
}

variable "aws_region" {
  type = "string"
  description = "The AWS region to create things in."
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

variable "stardog_home_volume_type" {
  type = "string"
  description = "The EBS storage type for the Stardog home volume"
  default = "io1"
}

variable "stardog_home_volume_size" {
  type = "string"
  description = "The size of the Stardog home volume in gigabytes."
  default = "10"
}

variable "stardog_home_volume_iops" {
  type = "string"
  description = "The IOPS to provision the Stardog home volume with"
  default = "500"
}
