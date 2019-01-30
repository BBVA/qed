output "private_ip" {
  value = "${aws_instance.inmemory-storage.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.inmemory-storage.public_ip}"
}
