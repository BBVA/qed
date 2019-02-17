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

module "qed" {
  source = "./modules/qed"

  name = "qed"
  count = 3
  instance_type = "t3.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
}

module "inmemory-storage" {
  source = "./modules/inmemory_storage"

  name = "inmemory-storage"
  instance_type = "t3.small"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
}

module "agent-publisher" {
  source = "./modules/qed"

  name = "agent-publisher"
  instance_type = "t3.small"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
  role = "publisher"

}

module "agent-monitor" {
  source = "./modules/qed"

  name = "agent-monitor"
  count = 2
  instance_type = "t3.small"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
  role = "monitor"
}

module "agent-auditor" {
  source = "./modules/qed"

  name = "agent-auditor"
  instance_type = "t3.small"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
  role = "auditor"

}

module "prometheus" {
  source = "./modules/prometheus"

  instance_type = "t3.medium"
  volume_size = "20"
  vpc_security_group_ids = "${module.prometheus_security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"

  # config = <<-CONFIG
  # global:
  #   scrape_interval:     15s
  #   evaluation_interval: 15s
  # scrape_configs:
  #   - job_name: 'prometheus'
  #     scrape_interval: 5s
  #     static_configs:
  #       - targets: ['localhost:9090']
  #   - job_name: 'Qed0-HostMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.leader.private_ip[0]}:9100']
  #   - job_name: 'Qed0-QedMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.leader.private_ip[0]}:8600']
  #   - job_name: 'Qed1-HostMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.follower-1.private_ip[0]}:9100']
  #   - job_name: 'Qed1-QedMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.follower-1.private_ip[0]}:8600']
  #   - job_name: 'Qed2-HostMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.follower-2.private_ip[0]}:9100']
  #   - job_name: 'Qed2-QedMetrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.follower-2.private_ip[0]}:8600']
  #   - job_name: 'Agent-Publisher-Metrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.agent-publisher.private_ip[0]}:18300']
  #   - job_name: 'Agent-Monitor-0-Metrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.agent-monitor.private_ip[0]}:18200']
  #   - job_name: 'Agent-Monitor-1-Metrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.agent-monitor.private_ip[1]}:18200']
  #   - job_name: 'Agent-Auditor-Metrics'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.agent-auditor.private_ip[0]}:18100']
  #   - job_name: 'riot'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.riot.private_ip}:9100']
  #   - job_name: 'inmemory-storage'
  #     scrape_interval: 10s
  #     static_configs:
  #       - targets: ['${module.inmemory-storage.private_ip}:18888']
  # CONFIG
}

module "riot" {
  source = "./modules/riot"

  instance_type = "t3.medium"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
  key_path = "${var.keypath}"
  endpoint =  "${module.qed.private_ip[0]}"
  num_requests = 10000000

}
