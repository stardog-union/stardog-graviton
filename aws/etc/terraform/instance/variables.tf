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
  default = "io1"
}

variable "root_volume_size" {
  type = "string"
  description = "The size of the root partition"
  default = "16"
}

variable "root_volume_iops" {
  type = "string"
  description = "The IOPS for the root volume"
  default = "800"
}

variable "custom_script" {
  type = "string"
  description = "A custom script to execute on stardog nodes"
  default = ""
}

variable "custom_zk_script" {
  type = "string"
  description = "A custom script to execute on zookeeper nodes"
  default = ""
}

# The remaining variables can be overridden by ENVs
variable "zk_unhealthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "5"
}

variable "zk_healthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "5"
}

variable "zk_health_timeout" {
  type = "string"
  description = "Zookeeper ELB health checker timeout"
  default = "5"
}

variable "zk_health_interval" {
  type = "string"
  description = "Zookeeper ELB health checker interval"
  default = "10"
}

variable "sd_unhealthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "10"
}

variable "sd_healthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "2"
}

variable "sd_health_timeout" {
  type = "string"
  description = "Zookeeper ELB health checker timeout"
  default = "10"
}

variable "sd_health_interval" {
  type = "string"
  description = "Zookeeper ELB health checker interval"
  default = "15"
}

variable "sd_internal_unhealthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "10"
}

variable "sd_internal_healthy_threshold" {
  type = "string"
  description = "Zookeeper ELB unhealthy threshold"
  default = "2"
}

variable "sd_internal_health_timeout" {
  type = "string"
  description = "Zookeeper ELB health checker timeout"
  default = "10"
}

variable "sd_internal_health_interval" {
  type = "string"
  description = "Zookeeper ELB health checker interval"
  default = "15"
}

variable "zk_health_grace_period" {
  type = "string"
  description = "Zookeeper ELB health checker grace period"
  default = "180"
}

variable "sd_health_grace_period" {
  type = "string"
  description = "Stardog ELB health checker grace period"
  default = "300"
}

variable "zk_tickTime" {
  type = "string"
  default = "3000"
}

variable "zk_initLimit" {
  type = "string"
  default = "10"
}

variable "zk_syncLimit" {
  type = "string"
  default = "5"
}

variable "zk_dataDir" {
  type = "string"
  default = "/var/zkdata"
}

variable "zk_dataLogDir" {
  type = "string"
  default = "/var/zkdata"
}

variable "zk_root_volume_type" {
  type = "string"
  default = "io1"
}

variable "zk_root_volume_size" {
  type = "string"
  default = "16"
}

variable "zk_root_volume_iops" {
  type = "string"
  default = "800"
}
