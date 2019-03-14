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
