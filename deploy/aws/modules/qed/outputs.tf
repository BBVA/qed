output "private_ip" {
  value = "${aws_instance.qed-server.*.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.qed-server.*.public_ip}"
}
