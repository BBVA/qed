
output "leader" {
  value = "${module.leader.public_ip}"
}

output "follower-1" {
  value = "${module.follower-1.public_ip}"
}
output "follower-2" {
  value = "${module.follower-2.public_ip}"
}
# output "inmemory-storage" {
#   value = "${module.inmemory-storage.public_ip}"
# }
# output "agent-publisher" {
#   value = "${module.agent-publisher.public_ip}"
# }
# output "agent-monitor" {
#   value = "${module.agent-monitor.public_ip}"
# }
# output "agent-auditor" {
#   value = "${module.agent-auditor.public_ip}"
# }

output "prometheus" {
  value = "${module.prometheus.public_ip}"
}
output "riot" {
  value = "${module.riot.public_ip}"
}

