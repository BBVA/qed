output "private_ip" {
  value = [
    for instance in aws_instance.qed-agent:
    instance.private_ip
  ]
}

output "public_ip" {
  value = [
    for instance in aws_instance.qed-agent:
    instance.public_ip
  ]
}
