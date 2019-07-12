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
resource null_resource "qed-base" {

  triggers {
    qed = "${format("%s",module.qed.private_ip)}"
    prometheus = "${module.prometheus.private_ip}"
    workload = "${module.workload.private_ip}"
    gateway = "${aws_internet_gateway.qed.id}"
    aws_route = "${aws_route.qed.id}"
    aws_vpc_dhcp_options = "${aws_vpc_dhcp_options.qed.id}"
    aws_vpc_dhcp_options_association = "${aws_vpc_dhcp_options_association.qed.id}"
    aws_cloudwatch_log_group = "${aws_cloudwatch_log_group.qed.name}"
    aws_iam_role = "${aws_iam_role.qed.id}"
    aws_flow_log = "${aws_flow_log.qed.id}"
    aws_iam_role_policy_attachmentCloudWatch = "${aws_iam_role_policy_attachment.CloudWatchLogsFullAccess-attach.role}"
    aws_iam_role_policy_attachmentQed = "${aws_iam_role_policy_attachment.qed.role}"
  }

}

resource null_resource "qed-full" {

  triggers {
    qed = "${format("%s",module.qed.private_ip)}"
    prometheus = "${module.prometheus.private_ip}"
    workload = "${module.workload.private_ip}"
    auditor = "${format("%s", module.agent-auditor.private_ip)}"
    monitor = "${format("%s", module.agent-monitor.private_ip)}"
    publisher = "${format("%s", module.agent-publisher.private_ip)}"
    storage = "${module.inmemory-storage.private_ip}"
    gateway = "${aws_internet_gateway.qed.id}"
    aws_route = "${aws_route.qed.id}"
    aws_vpc_dhcp_options = "${aws_vpc_dhcp_options.qed.id}"
    aws_vpc_dhcp_options_association = "${aws_vpc_dhcp_options_association.qed.id}"
    aws_cloudwatch_log_group = "${aws_cloudwatch_log_group.qed.name}"
    aws_iam_role = "${aws_iam_role.qed.id}"
    aws_flow_log = "${aws_flow_log.qed.id}"
    aws_iam_role_policy_attachmentCloudWatch = "${aws_iam_role_policy_attachment.CloudWatchLogsFullAccess-attach.role}"
    aws_iam_role_policy_attachmentQed = "${aws_iam_role_policy_attachment.qed.role}"
  }

}
