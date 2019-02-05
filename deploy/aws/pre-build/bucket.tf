terraform {
  required_version = ">= 0.11.11"
}

provider "aws" {
  version = ">= 1.56.0, < 2.0"
  region = "eu-west-1"
  profile = "${var.aws_profile}"
}

resource "aws_s3_bucket" "terraform-qed-cluster" {
    bucket = "terraform-qed-cluster"
 
    versioning {
      enabled = true
    }
 
    lifecycle {
      prevent_destroy = true
    }
 
    tags {
      Name = "S3 Remote Terraform State Store"
    }      
}
