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

output "id" {
  description = "List of IDs of instances"
  value       = ["${module.ec2.id}"]
}

output "public_dns" {
  description = "List of public DNS names assigned to the instances"
  value       = ["${module.ec2.public_dns}"]
}

output "instance_id" {
  description = "EC2 instance ID"
  value       = "${module.ec2.id[0]}"
}

output "instance_public_dns" {
  description = "Public DNS name assigned to the EC2 instance"
  value       = "${module.ec2.public_dns[0]}"
}

output "spartan_public_dns" {
  description = "Public DNS name assigned to the EC2 Spartan"
  value       = "${module.ec2-spartan.public_dns[0]}"
}

output "credit_specification" {
  description = "Credit specification of EC2 instance (empty list for not t2 instance types)"
  value       = "${module.ec2.credit_specification}"
}