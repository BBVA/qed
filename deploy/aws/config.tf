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

terraform {
  required_version = ">= 0.11.11"
}

provider "aws" {
  version = ">= 1.56.0, < 2.0"
  profile = "${var.aws_profile}"
}

provider "http" {
  version = ">= 1.0.1, < 2.0"
}

provider "null" {
  version = "~> 2.0"
}
