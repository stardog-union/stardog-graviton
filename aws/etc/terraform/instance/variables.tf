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
  default = "t2.medium"
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
      "us-west-1b"]
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
    ],
    eu-west-2 = [
      "eu-west-2a",
      "eu-west-2b"
    ],
    ap-northeast-1 = [
      "ap-northeast-1a",
      "ap-northeast-1c"
    ],
    ap-northeast-2 = [
      "ap-northeast-2a",
      "ap-northeast-2c"
    ],
    ap-southeast-1 = [
      "ap-southeast-1a",
      "ap-southeast-1b"
    ],
    ap-southeast-2 = [
      "ap-southeast-2a",
      "ap-southeast-2b"
    ],
    sa-east-1 = [
      "sa-east-1a",
      "sa-east-1c"
    ],
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

variable "elb_idle_timeout" {
  type = "string"
  description = "The amount of time in seconds that a connection to a stardog load balancer can be idle."
  default = "300"
}

variable "custom_properties_data" {
  type = "string"
  description = "The custom data to add to the stardog properties"
  default = ""
}

variable "custom_log4j_data" {
  type = "string"
  description = "The custom log4j file"
  default = ""
}

variable "environment_variables" {
  type = "string"
  description = "The environment variables to inject into the stardog script"
  default = ""
}

variable "stardog_start_opts" {
  type = "string"
  description = "Options passed to stardog-admin server start"
  default = ""
}

variable "root_volume_type" {
  type = "string"
  description = "The type of volume to use for the root partition"
  default = "standard"
}

variable "root_volume_size" {
  type = "string"
  description = "The size of the root partition"
  default = "16g"
}

variable "custom_script" {
  type = "string"
  description = "A custom script to execute on stardog nodes"
  default = ""
}
