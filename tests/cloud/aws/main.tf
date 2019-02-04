/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

provider "aws" {
  region = "${var.region}"
  profile = "${var.profile}"
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_key_pair" "qed-benchmark" {
  key_name   = "qed-benchmark"
  public_key = "${file("~/.ssh/id_rsa.pub")}"
}

data "aws_subnet_ids" "all" {
  vpc_id = "${data.aws_vpc.default.id}"
}

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name = "name"

    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"

    values = [
      "amazon",
    ]
  }
}

data "http" "ip" {
  url = "http://icanhazip.com"
}

module "security_group" {
  source = "terraform-aws-modules/security-group/aws"

  name        = "qed-benchmark"
  description = "Security group for QED benchmark usage"
  vpc_id      = "${data.aws_vpc.default.id}"

  ingress_cidr_blocks = ["${chomp(data.http.ip.body)}/32"]
  ingress_rules       = ["http-8800-tcp", "all-icmp", "ssh-tcp" ]
  egress_rules        = ["all-all"]
}

resource "aws_security_group_rule" "allow_profiling" {
  type            = "ingress"
  from_port       = 6060
  to_port         = 6060
  protocol        = "tcp"
  cidr_blocks     = ["${chomp(data.http.ip.body)}/32"]

  security_group_id = "${module.security_group.this_security_group_id}"
}

resource "aws_security_group_rule" "allow_cluster_comm" {
  type            = "ingress"
  from_port       = 0
  to_port         = 65535
  protocol        = "tcp"
  source_security_group_id  = "${module.security_group.this_security_group_id}"

  security_group_id = "${module.security_group.this_security_group_id}"
}

module "ec2" {
  source = "terraform-aws-modules/ec2-instance/aws"

  name                        = "qed-benchmark"
  ami                         = "${data.aws_ami.amazon_linux.id}"
  instance_type               = "${var.flavour}"
  instance_count              = "${var.cluster_size}"
  subnet_id                   = "${element(data.aws_subnet_ids.all.ids, 0)}"
  vpc_security_group_ids      = ["${module.security_group.this_security_group_id}"]
  associate_public_ip_address = true
  key_name                    = "${aws_key_pair.qed-benchmark.key_name}"

  root_block_device = [{
    volume_type = "gp2"
    volume_size = "${var.volume_size}"
    delete_on_termination = true
  }]

}

// Bring up the stress instance.
module "ec2-spartan" {
  source = "terraform-aws-modules/ec2-instance/aws"

  name                        = "qed-benchmark-spartan"
  ami                         = "${data.aws_ami.amazon_linux.id}"
  instance_type               = "${var.flavour}"
  subnet_id                   = "${element(data.aws_subnet_ids.all.ids, 0)}"
  vpc_security_group_ids      = ["${module.security_group.this_security_group_id}"]
  associate_public_ip_address = true
  key_name                    = "${aws_key_pair.qed-benchmark.key_name}"

  root_block_device = [{
    volume_type = "gp2"
    volume_size = "${var.volume_size}"
    delete_on_termination = true
  }]

}

// Build qed and outputs a single binary file
resource "null_resource" "build-qed" {
  provisioner "local-exec" {
    command = "go build -o to_upload/qed ../../../"
  }

  depends_on = ["module.ec2"]

}

# Template for initial configuration bash script
 resource "template_dir" "gen-single-node-config" {
   count = "${var.cluster_size == 1 ? 1:0}"

   source_dir = "templates"
   destination_dir = "to_upload/rendered"

   vars {
     master_address  = "${module.ec2.private_ip[0]}"
     slave01_address = ""
     slave02_address = ""
   }

   depends_on = ["module.ec2", "null_resource.build-qed"]
 }

# Template for initial configuration bash script
 resource "template_dir" "gen-multi-node-config" {
   count = "${var.cluster_size > 1 ? 1:0}"

   source_dir = "templates"
   destination_dir = "to_upload/rendered"

   vars {
     master_address = "${module.ec2.private_ip[0]}"
     slave01_address = "${module.ec2.private_ip[1]}"
     slave02_address = "${module.ec2.private_ip[2]}"
   }

   depends_on = ["module.ec2", "null_resource.build-qed"]
 }

// Copies qed binary and bench tools to out EC2 instance using SSH
resource "null_resource" "copy-qed-to-nodes" {
  count       = "${var.cluster_size}"

  provisioner "file" {
    source      = "to_upload"
    destination = "/tmp"
    
    connection {
      host        = "${element(module.ec2.public_ip, count.index)}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }
  }

  depends_on = ["null_resource.build-qed", "module.ec2", "template_dir.gen-single-node-config", "template_dir.gen-multi-node-config"]

}

// Copies qed binary and bench tools to out EC2 instance using SSH
resource "null_resource" "copy-qed-to-spartan" {

  provisioner "file" {
    source      = "to_upload"
    destination = "/tmp"
    
    connection {
      host        = "${module.ec2-spartan.public_ip[0]}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }
  }

  depends_on = ["null_resource.build-qed", "module.ec2", "template_dir.gen-single-node-config", "template_dir.gen-multi-node-config"]

}

resource "null_resource" "install-tools-to-spartan" {
  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/to_upload/install-tools /tmp/to_upload/rendered/stress-throughput-60s /tmp/to_upload/qed",
      "/tmp/to_upload/install-tools",
    ]
   
    connection {
      host        = "${module.ec2-spartan.public_ip}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }

  }

  depends_on = ["null_resource.copy-qed-to-spartan"]

}

resource "null_resource" "start-master" {

  provisioner "remote-exec" {
    inline = [
      "find /tmp/to_upload -type f -exec chmod a+x {} \\;",
      "/tmp/to_upload/rendered/start_master",
    ]

    connection {
      host        = "${module.ec2.public_ip[0]}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }
  }

  depends_on = ["null_resource.copy-qed-to-nodes"]

}

resource "null_resource" "start-slave" {

  count = "${var.cluster_size > 1 ? var.cluster_size - 1 : 0}"

  provisioner "remote-exec" {
    inline = [
      "find /tmp/to_upload -type f -exec chmod a+x {} \\;",
      "/tmp/to_upload/rendered/start_${element(var.slave_prefix, count.index)}",
    ]

    connection {
      host        = "${element(module.ec2.public_ip, count.index+1)}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }
  }

  depends_on = ["null_resource.start-master"]

}

resource "null_resource" "run-benchmarks" {
  provisioner "remote-exec" {
    inline = [
      "/tmp/to_upload/rendered/stress-throughput-60s",
    ]

    connection {
      host        = "${module.ec2-spartan.public_ip}"
      type        = "ssh"
      private_key = "${file("~/.ssh/id_rsa")}"
      user        = "ec2-user"
      timeout     = "5m"
    }
  }

   depends_on = ["null_resource.install-tools-to-spartan", "null_resource.start-master", "null_resource.start-slave"]

}
