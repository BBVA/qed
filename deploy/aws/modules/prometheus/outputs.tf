output "private_ip" {
  value = [
    for instance in aws_instance.prometheus:
    instance.private_ip
  ]
}

output "public_ip" {
  value = [
    for instance in aws_instance.prometheus:
    instance.public_ip
  ]
}
