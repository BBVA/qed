terraform {
  required_version = ">= 0.11.11"
}

provider "aws" {
  version = ">= 1.56.0, < 2.0"
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

  tags {
    Name = "S3 Remote Terraform State Store"
  }
}
