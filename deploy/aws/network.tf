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

data "aws_caller_identity" "current" {
}

resource "aws_vpc" "qed" {
  enable_dns_hostnames = true
  cidr_block           = var.vpc_cidr

  tags = {
    Name = "QED-${terraform.workspace}"
  }
}

resource "aws_subnet" "qed" {
  vpc_id                  = aws_vpc.qed.id
  cidr_block              = var.public_subnet_cidr
  map_public_ip_on_launch = true

  tags = {
    Name = "QED-${terraform.workspace}"
  }
}

resource "aws_internet_gateway" "qed" {
  vpc_id = aws_vpc.qed.id

  tags = {
    Name = "QED-${terraform.workspace}"
  }
}

resource "aws_route" "qed" {
  route_table_id         = aws_vpc.qed.default_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.qed.id
}

resource "aws_vpc_dhcp_options" "qed" {
  domain_name         = "service.qed"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    Name = "QED-${terraform.workspace}"
  }
}

resource "aws_vpc_dhcp_options_association" "qed" {
  vpc_id          = aws_vpc.qed.id
  dhcp_options_id = aws_vpc_dhcp_options.qed.id
}

# https://github.com/hashicorp/terraform/issues/14750
# CloudWatch LogGroup is not removed because there are still live Log Flow streams.
resource "aws_cloudwatch_log_group" "qed" {
  name = "qed-${terraform.workspace}"
}

resource "aws_iam_role" "qed" {
  name                 = "qed-${terraform.workspace}"
  permissions_boundary = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/CoderPermissionsBoundaries"
  assume_role_policy   = <<EOF
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
  role       = aws_iam_role.qed.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_flow_log" "qed" {
  log_destination = aws_cloudwatch_log_group.qed.arn
  iam_role_arn    = aws_iam_role.qed.arn
  vpc_id          = aws_vpc.qed.id
  traffic_type    = "ALL"
}

resource "aws_key_pair" "qed" {
  key_name   = "qed-${terraform.workspace}"
  public_key = file("${var.keypath}.pub")
}

# Create the Security Group
resource "aws_security_group" "qed" {
  vpc_id      = aws_vpc.qed.id
  name        = "qed-${terraform.workspace}"
  description = "Security group for QED usage"
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 8800
    to_port     = 8800
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 8888
    to_port     = 8888
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 8600
    to_port     = 8600
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 6060
    to_port     = 6060
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 7700
    to_port     = 7700
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port   = 9100
    to_port     = 9100
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }
  ingress {
    from_port       = 0
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.prometheus.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = {
    Name      = "qed-${terraform.workspace}"
    Workspace = terraform.workspace
  }
}

resource "aws_security_group" "prometheus" {
  vpc_id      = aws_vpc.qed.id
  name        = "prometheus-${terraform.workspace}"
  description = "Security group for QED usage"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }

  # Prometheus metrics
  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }

  # Grafana
  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  }
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = {
    Name      = "prometheus-${terraform.workspace}"
    Workspace = terraform.workspace
  }
}

