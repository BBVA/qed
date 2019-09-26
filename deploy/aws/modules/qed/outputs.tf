output "private_ip" {
  value = [
    for instance in aws_instance.qed-server:
    instance.private_ip
  ]
}

output "public_ip" {
  value = [
    for instance in aws_instance.qed-server:
    instance.public_ip
  ]
}
