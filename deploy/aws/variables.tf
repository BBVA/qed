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
variable "aws_profile" {
  default = "bbva-labs"
}

variable "keypath" {
  default = "~/.ssh/id_rsa-qed"
}

variable "vpc_cidr" {
  description = "CIDR of the VPC as a whole"
  default     = "172.31.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR of the public subnet"
  default     = "172.31.1.0/24"
}

