#  Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.
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

data "aws_iam_policy_document" "CloudWatchLogsFullAccess-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "CloudWatchLogsFullAccess" {
  name               = "CloudWatchLogsFullAccess-${terraform.workspace}"
  permissions_boundary = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/PermissionsBoundariesBBVA"
  assume_role_policy = "${data.aws_iam_policy_document.CloudWatchLogsFullAccess-assume-role-policy.json}"
}

resource "aws_iam_role_policy_attachment" "CloudWatchLogsFullAccess-attach" {
  role       = "${aws_iam_role.CloudWatchLogsFullAccess.name}"
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_iam_instance_profile" "qed-profile" {
  name = "qed-profile-${terraform.workspace}"
  role = "${aws_iam_role.CloudWatchLogsFullAccess.name}"
}

module "qed" {
  source = "./modules/qed"
  count  = 3

  name                   = "qed"
  instance_type          = "z1d.xlarge"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}
module "inmemory-storage" {
  source = "./modules/inmemory_storage"

  name                   = "inmemory-storage"
  instance_type          = "r5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}

module "agent-publisher" {
  source = "./modules/agent"
  count  = 1

  role                   = "publisher"
  name                   = "agent-publisher"
  instance_type          = "m5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}

module "agent-monitor" {
  source = "./modules/agent"
  count  = 1

  role                   = "monitor"
  name                   = "agent-monitor"
  instance_type          = "m5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}

module "agent-auditor" {
  source = "./modules/agent"
  count  = 1

  role                   = "auditor"
  name                   = "agent-auditor"
  instance_type          = "m5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}

module "prometheus" {
  source = "./modules/prometheus"

  instance_type          = "m5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.prometheus.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
}

module "workload" {
  source = "./modules/workload"

  instance_type          = "m5.large"
  iam_instance_profile   = "${aws_iam_instance_profile.qed-profile.name}"
  volume_size            = "20"
  vpc_security_group_ids = "${aws_security_group.qed.id}"
  subnet_id              = "${aws_subnet.qed.id}"
  key_name               = "${aws_key_pair.qed.key_name}"
  key_path               = "${var.keypath}"
  endpoint               = "${module.qed.private_ip[0]}"
  num_requests           = 10000000
}
