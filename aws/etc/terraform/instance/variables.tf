variable "zk_instance_type" {
  type = "string"
  description = "The AWS instance type for zookeeper nodes."
}

variable "stardog_instance_type" {
  type = "string"
  description = "The AWS instance type for stardog."
}

variable "bastion_instance_type" {
  type = "string"
  default = "m3.medium"
  description = "The AWS instance type for stardog."
}

variable "aws_key_name" {
  type = "string"
  description = "The AWS key name"
}

variable "zookeeper_size" {
  type = "string"
  description = "The number of zookeeper nodes to use (must be odd and greater than 1)"
}

variable "stardog_size" {
  type = "string"
  description = "The number of stardog nodes to use (must be odd and greater than 1)"
}

variable "baseami" {
  type = "string"
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
      "us-west-2c"]
  }
}

variable "deployment_name" {
  type = "string"
  description = "A string that is unique to this stardog data in a given account and resource"
}

variable "version" {
  type = "string"
  description = "The version of stardog to launch."
}

variable "internal_network" {
  type = "string"
  description = "The internal subnet for the virtual appliance."
  default = "10.0.0.0/16"
}

variable "http_subnet" {
  type = "string"
  description = "The internal subnet for the virtual appliance."
}

variable "custom_stardog_properties" {
  type = "string"
  description = "Custom entries for stardog.properties"
  default = ""
}

variable "external_protocol" {
  type = "string"
  description = "The protocol to use on the load balancer.  Must be http or https."
  default = "http"
}

variable "ssl_cert_arn" {
  type = "string"
  description = "The ARN to the cert."
  default = ""
}


