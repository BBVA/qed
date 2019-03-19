/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

data "http" "ip" {
  url = "http://icanhazip.com"
}

resource "aws_vpc" "qed" {
  enable_dns_hostnames = true
  cidr_block           = "${var.vpc_cidr}"

  tags = {
    Name = "QED"
  }
}

resource "aws_subnet" "qed" {
  vpc_id                  = "${aws_vpc.qed.id}"
  cidr_block              = "${var.public_subnet_cidr}"
  map_public_ip_on_launch = true

  tags = {
    Name = "QED"
  }
}

resource "aws_internet_gateway" "qed" {
  vpc_id = "${aws_vpc.qed.id}"

  tags = {
    Name = "QED"
  }
}

resource "aws_route" "qed" {
  route_table_id         = "${aws_vpc.qed.default_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.qed.id}"
}

resource "aws_vpc_dhcp_options" "qed" {
  domain_name         = "service.qed"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    Name = "QED"
  }
}

resource "aws_vpc_dhcp_options_association" "qed" {
  vpc_id          = "${aws_vpc.qed.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.qed.id}"
}

data "aws_cloudwatch_log_group" "qed" {
  name = "qed"
}

resource "aws_iam_role" "qed" {
  name = "qed"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "vpc-flow-logs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "qed" {
  role       = "${aws_iam_role.qed.name}"
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_flow_log" "qed" {
  log_destination = "${data.aws_cloudwatch_log_group.qed.arn}"
  iam_role_arn    = "${aws_iam_role.qed.arn}"
  vpc_id          = "${aws_vpc.qed.id}"
  traffic_type    = "ALL"
}

resource "aws_key_pair" "qed" {
  key_name   = "qed"
  public_key = "${file("${var.keypath}.pub")}"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "2.11.0"

  name        = "qed"
  description = "Security group for QED usage"
  vpc_id      = "${aws_vpc.qed.id}"

  egress_rules = ["all-all"]

  ingress_cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  ingress_rules       = ["all-icmp", "ssh-tcp"]

  ingress_with_cidr_blocks = [
    {
      from_port   = 8800
      to_port     = 8800
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 8888
      to_port     = 8888
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 8600
      to_port     = 8600
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 6060
      to_port     = 6060
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 7700
      to_port     = 7700
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 9100
      to_port     = 9100
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
  ]

  computed_ingress_with_source_security_group_id = [
    {
      from_port                = 0
      to_port                  = 65535
      protocol                 = "tcp"
      source_security_group_id = "${module.security_group.this_security_group_id}"
    },
    {
      from_port                = 0
      to_port                  = 65535
      protocol                 = "tcp"
      source_security_group_id = "${module.prometheus_security_group.this_security_group_id}"
    },
  ]

  number_of_computed_ingress_with_source_security_group_id = 2
}

module "prometheus_security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "2.11.0"

  name        = "prometheus"
  description = "Security group for Prometheus/Grafana usage"
  vpc_id      = "${aws_vpc.qed.id}"

  egress_rules = ["all-all"]

  ingress_cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  ingress_rules       = ["all-icmp", "ssh-tcp"]

  ingress_with_cidr_blocks = [
    {
      from_port   = 9090                             # prometheus metrics
      to_port     = 9090
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
    {
      from_port   = 3000                             # graphana
      to_port     = 3000
      protocol    = "tcp"
      cidr_blocks = "${chomp(data.http.ip.body)}/32"
    },
  ]

  computed_ingress_with_source_security_group_id = [
    {
      from_port                = 0
      to_port                  = 65535
      protocol                 = "tcp"
      source_security_group_id = "${module.security_group.this_security_group_id}"
    },
  ]

  number_of_computed_ingress_with_source_security_group_id = 1
}
