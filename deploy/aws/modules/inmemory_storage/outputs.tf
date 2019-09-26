output "private_ip" {
  value = [
    for instance in aws_instance.inmemory-storage:
    instance.private_ip
  ]
}

output "public_ip" {
  value = [
    for instance in aws_instance.inmemory-storage:
    instance.public_ip
  ]
}
