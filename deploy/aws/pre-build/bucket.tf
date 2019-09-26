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
terraform {
  required_version = ">= 0.12.9"
}

provider "aws" {
  version = ">= 2.7.0"
  region  = "eu-west-1"
  profile = "${var.aws_profile}"
}

resource "aws_kms_key" "bucket-key" {
  description             = "This key is used to encrypt bucket objects"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "terraform-qed-cluster" {
  bucket = "terraform-qed-cluster"
  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = true
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = "${aws_kms_key.bucket-key.arn}"
        sse_algorithm     = "aws:kms"
      }
    }
  }

  tags = {
    Name = "S3 Remote Terraform State Store"
  }
}
