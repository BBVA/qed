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

variable "instance_type" {}

variable "volume_size" {}

variable "vpc_security_group_ids" {}

variable "subnet_id" {}

variable "key_name" {}

variable "key_path" {}
variable "role" {
  default = "promethreus"
}
variable "path" {
  default = "/home/ec2-user/prometheus"
}
