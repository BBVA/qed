output "private_ip" {
  value = "${aws_instance.qed-agent.*.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.qed-agent.*.public_ip}"
}
