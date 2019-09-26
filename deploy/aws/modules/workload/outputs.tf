output "private_ip" {
  value = [
    for instance in aws_instance.workload:
    instance.private_ip
  ]
}

output "public_ip" {
    value = [
    for instance in aws_instance.workload:
    instance.public_ip
  ]
}
