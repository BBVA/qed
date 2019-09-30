#!/usr/bin/env bash
#  Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# TF_STATE must be the realtive or absolute path to the terraform 
# project directory. This is required due a lack decent multi version
# tfstate support of the terraform-inventory library.
# Issue: https://github.com/adammck/terraform-inventory/pull/114

export TF_STATE=../aws

if ! which terraform-inventory
then
    echo -e "Please install terraform-inventory in your GOBIN \ngo get github.com/adammck/terraform-inventory"
    exit 1
fi

cd provision

if [ -z "$@" ];
then
    ansible-playbook --inventory-file=$(which terraform-inventory) --private-key ~/.ssh/id_rsa-qed main.yml -f 10
else
    echo "Using custom Ansible Playbook command."
    ansible-playbook --inventory-file=$(which terraform-inventory) --private-key ~/.ssh/id_rsa-qed -f 10 "$@"
fi
