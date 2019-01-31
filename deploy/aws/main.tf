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
    command = "bash config_build.sh"
    working_dir = "/config_file"
  }
}

module "leader" {
  source = "./modules/qed"

  name = "leader"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  server:
    node_id: "leader"
    addr:
      http: ":8080"
      mgmt: ":8090"
      raft: ":9000"
      gossip: ":9100"
  CONFIG

}

module "follower-1" {
  source = "./modules/qed"

  name = "follower-1"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  server:
    node_id: "follower-1"
    addr:
      http: ":8080"
      mgmt: ":8090"
      raft: ":9000"
      gossip: ":9100"
      raft_join:
       - "${module.leader.private_ip}:9000"
      gossip_join:
        - "${module.leader.private_ip}:9100"
  CONFIG

}

module "follower-2" {
  source = "./modules/qed"

  name = "follower-2"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  server:
    node_id: "follower-2"
    addr:
      http: ":8080"
      mgmt: ":8090"
      raft: ":9000"
      gossip: ":9100"
      raft_join:
       - "${module.leader.private_ip}:9000"
      gossip_join:
        - "${module.leader.private_ip}:9100"
  CONFIG
}

module "inmemory-storage" {
  source = "./modules/inmemory_storage"

  name = "inmemory-storage"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"
}

module "agent-publisher" {
  source = "./modules/qed"

  name = "agent-publisher"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  command="agent"
  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  agent:
    node: "publisher"
    bind: ":9300"
    advertise: ""
    join:
      - "${module.leader.private_ip}:9100"
    server_urls:
      - "${module.leader.private_ip}:8080"
    alert_urls:
      - "${module.inmemory-storage.private_ip}:8888"
    snapshots_store_urls:
      - "${module.inmemory-storage.private_ip}:8888"
  CONFIG
}


module "agent-monitor" {
  source = "./modules/qed"

  name = "agent-monitor"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  command="agent"
  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  agent:
    node: "monitor"
    bind: ":9200"
    advertise: ""
    join:
      - "${module.leader.private_ip}:9100"
    server_urls:
      - "${module.leader.private_ip}:8080"
    alert_urls:
      - "${module.inmemory-storage.private_ip}:8888"
    snapshots_store_urls:
      - "${module.inmemory-storage.private_ip}:8888"
  CONFIG
}

module "agent-auditor" {
  source = "./modules/qed"

  name = "agent-auditor"
  instance_type = "t2.2xlarge"
  volume_size = "20"
  vpc_security_group_ids = "${module.security_group.this_security_group_id}"
  subnet_id = "${element(data.aws_subnet_ids.all.ids, 0)}"
  key_name = "${aws_key_pair.qed.key_name}"

  command="agent"
  config = <<-CONFIG
  ---
  api_key: "terraform_qed"
  path: "/var/tmp/qed/"

  agent:
    node: "auditor"
    bind: ":9100"
    advertise: ""
    join:
      - "${module.leader.private_ip}:9100"
    server_urls:
      - "${module.leader.private_ip}:8080"
    alert_urls:
      - "${module.inmemory-storage.private_ip}:8888"
    snapshots_store_urls:
      - "${module.inmemory-storage.private_ip}:8888"
  CONFIG
}
