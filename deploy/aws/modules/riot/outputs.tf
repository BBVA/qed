output "private_ip" {
  value = "${aws_instance.riot.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.riot.public_ip}"
}
