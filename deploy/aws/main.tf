#  Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

resource "null_resource" "prebuild" {
  provisioner "local-exec" {
    command = "bash ./config_build.sh"
  }
}

module "leader" {
  source = "./modules/qed"

  name = "qed0"
  instance_type = "t3.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"
  server:
    node_id: "qed0"
    addr:
      http: ":8800"
      mgmt: ":8700"
      raft: ":8500"
      gossip: ":8400"
  CONFIG

}

module "follower-1" {
  source = "./modules/qed"

  name = "qed1"
  instance_type = "t3.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"
  server:
    node_id: "qed1"
    addr:
      http: ":8800"
      mgmt: ":8700"
      raft: ":8500"
      gossip: ":8400"
      raft_join:
       - "${module.leader.private_ip}:8700"
      gossip_join:
        - "${module.leader.private_ip}:8400"
  CONFIG

}

module "follower-2" {
  source = "./modules/qed"

  name = "qed2"
  instance_type = "t3.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"
  server:
    node_id: "qed2"
    addr:
      http: ":8800"
      mgmt: ":8700"
      raft: ":8500"
      gossip: ":8400"
      raft_join:
       - "${module.leader.private_ip}:8700"
      gossip_join:
        - "${module.leader.private_ip}:8400"
  CONFIG
}

# module "inmemory-storage" {
#   source = "./modules/inmemory_storage"
#
#   name = "inmemory-storage"
#   instance_type = "t3.small"
#   volume_size = "20"
#   vpc_security_group_ids = "${module.security_group.this_security_group_id}"
#   subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
#   key_name = "${aws_key_pair.qed.key_name}"
# }

# module "agent-publisher" {
#   source = "./modules/qed"
#
#   name = "agent-publisher"
#   instance_type = "t3.small"
#   volume_size = "20"
#   vpc_security_group_ids = "${module.security_group.this_security_group_id}"
#   subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
#   key_name = "${aws_key_pair.qed.key_name}"
#
#   command="agent publisher"
#   config = <<-CONFIG
#   ---
#   api_key: "terraform_qed"
#   path: "/var/tmp/qed/"
#   agent:
#     node: "publisher"
#     bind: ":8300"
#     advertise: ""
#     join:
#       - "${module.leader.private_ip}:8400"
#     server_urls:
#       - "${module.leader.private_ip}:8800"
#     alert_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#     snapshots_store_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#   CONFIG
# }


# module "agent-monitor" {
#   source = "./modules/qed"
#
#   name = "agent-monitor"
#   instance_type = "t3.small"
#   volume_size = "20"
#   vpc_security_group_ids = "${module.security_group.this_security_group_id}"
#   subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
#   key_name = "${aws_key_pair.qed.key_name}"
#
#   command="agent monitor"
#   config = <<-CONFIG
#   ---
#   api_key: "terraform_qed"
#   path: "/var/tmp/qed/"
#   agent:
#     node: "monitor"
#     bind: ":8200"
#     advertise: ""
#     join:
#       - "${module.leader.private_ip}:8400"
#     server_urls:
#       - "${module.leader.private_ip}:8800"
#     alert_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#     snapshots_store_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#   CONFIG
# }

# module "agent-auditor" {
#   source = "./modules/qed"
#
#   name = "agent-auditor"
#   instance_type = "t2.2xlarge"
#   volume_size = "20"
#   vpc_security_group_ids = "${module.security_group.this_security_group_id}"
#   subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
#   key_name = "${aws_key_pair.qed.key_name}"
#
#   command="agent auditor"
#   config = <<-CONFIG
#   ---
#   api_key: "terraform_qed"
#   path: "/var/tmp/qed/"
#   agent:
#     node: "auditor"
#     bind: ":8100"
#     advertise: ""
#     join:
#       - "${module.leader.private_ip}:8400"
#     server_urls:
#       - "${module.leader.private_ip}:8800"
#     alert_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#     snapshots_store_urls:
#       - "${module.inmemory-storage.private_ip}:8888"
#   CONFIG
# }

module "prometheus" {
  source = "./modules/prometheus"

  instance_type = "t3.medium"
  volume_size = "20"
  vpc_security_group_ids = "${module.prometheus_security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  global:
    scrape_interval:     15s
    evaluation_interval: 15s
  scrape_configs:
    - job_name: 'prometheus'
      scrape_interval: 5s
      static_configs:
        - targets: ['localhost:9090']
    - job_name: 'Qed0-HostMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.leader.private_ip}:9100', ]
    - job_name: 'Qed0-QedMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.leader.private_ip}:8600']
    - job_name: 'Qed1-HostMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.follower-1.private_ip}:9100']
    - job_name: 'Qed1-QedMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.follower-1.private_ip}:8600']
    - job_name: 'Qed2-HostMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.follower-2.private_ip}:9100']
    - job_name: 'Qed2-QedMetrics'
      scrape_interval: 10s
      static_configs:
        - targets: ['${module.follower-2.private_ip}:8600']
  CONFIG
}
