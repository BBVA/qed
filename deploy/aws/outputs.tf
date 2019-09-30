output "qed" {
  value = module.qed.public_ip
}

output "qed-private" {
  value = module.qed.private_ip
}

output "prometheus" {
  value = module.prometheus.public_ip
}

output "workload" {
  value = module.workload.public_ip
}

output "inmemory-storage" {
  value = module.inmemory-storage.public_ip
}

output "agent-publisher" {
  value = module.agent-publisher.public_ip
}

output "agent-monitor" {
  value = module.agent-monitor.public_ip
}

output "agent-auditor" {
  value = module.agent-auditor.public_ip
}

