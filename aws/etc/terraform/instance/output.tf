output "stardog_contact" {
  #  value = ["${aws_elb.stardog.*.dns_name}"]
  value = "${aws_elb.stardog.dns_name}"
}

output "stardog_internal_contact" {
  #  value = ["${aws_elb.stardog.*.dns_name}"]
  value = "${aws_elb.stardoginternal.dns_name}"
}

output "bastion_contact" {
  value = "${aws_elb.bastion.dns_name}"
}

output "zookeeper_nodes" {
  value = ["${aws_elb.zookeeper.*.dns_name}"]
}

output "stardog_node_ip" {
  value = ["${aws_elb.stardoginternal.instances}"]
}