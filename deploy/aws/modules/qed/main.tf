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

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
}

resource "aws_instance" "qed-server" {
  count                       = "${var.count}"
  ami                         = "${data.aws_ami.amazon_linux.id}"
  instance_type               = "${var.instance_type}"

  vpc_security_group_ids      = ["${var.vpc_security_group_ids}"]
  subnet_id                   = "${var.subnet_id}"
  associate_public_ip_address = true
  key_name                    = "${var.key_name}"

  root_block_device = [{
    volume_type = "gp2"
    volume_size = "${var.volume_size}"
  }]

  tags {
    Name = "${format("${var.name}-%01d", count.index)}"
  }


  provisioner "file" {
      source     = "./config_files"
      destination = "${var.path}"

      connection {
        user = "ec2-user"
        private_key = "${file("${var.key_path}")}"
      }
  }

  provisioner "file" {
      content     = "${var.config}"
      destination = "${var.path}/config.yml"

      connection {
        user = "ec2-user"
        private_key = "${file("${var.key_path}")}"
      }
  }

  user_data = <<-DATA
  #!/bin/bash

  echo "Install and enable AWS CloudWatch"
  aws configure set region eu-west-1 
  sudo yum install -y awslogs
  sudo sed -i "s/us-east-1/eu-west-1/g" /etc/awslogs/awscli.conf
sudo cat << EOF >> /etc/awslogs/awslogs.conf

[/var/tmp/qed]
datetime_format = %b %d %H:%M:%S
file = /var/tmp/qed/qed.log
buffer_duration = 5000
log_stream_name = ${var.name}
initial_position = start_of_file
log_group_name = qed
EOF

  sudo service awslogs start

  while [ ! -f ${var.path}/qed ]; do
    sleep 1 # INFO: wait until binary exists
  done

  while [ ! -f ${var.path}/config.yml ]; do
    sleep 1 # INFO: wait until file exists
  done

  while [ `lsof ${var.path}/* | wc -l` -gt 0 ]; do
    sleep 1 # INFO: prevents Error of `text file busy`
  done

  chmod +x ${var.path}/qed
  chmod +x ${var.path}/node_exporter

  ${var.path}/node_exporter &
  
  # TONIGHT WE DINE IN HELL
  sed -i "s/MYIP/$(ifconfig | awk '/inet addr/{print substr($2,6)}' | head -1)/g" ${var.path}/config.yml
  ${var.path}/qed ${var.command}

  DATA
}
