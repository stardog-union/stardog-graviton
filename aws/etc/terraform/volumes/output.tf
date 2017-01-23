output "volumes" {
  value = ["${aws_ebs_volume.stardog_data.*.id}"]
}
