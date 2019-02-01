output "private_ip" {
  value = "${aws_instance.prometheus.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.prometheus.public_ip}"
}
