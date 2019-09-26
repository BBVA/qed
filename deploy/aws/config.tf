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

# Bucket config must be here: https://github.com/hashicorp/terraform/issues/13589
terraform {
  required_version = ">= 0.12.0"

  backend "s3" {
    bucket = "terraform-qed-cluster"
    key    = "terraform.tfstate"
    region = "eu-west-1"
  }
}

provider "aws" {
  version = ">= 2.7.0"
  profile = var.aws_profile
}

provider "http" {
  version = ">= 1.0.1, < 2.0"
}

provider "null" {
  version = "~> 2.0"
}

