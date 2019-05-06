#!/usr/bin/env bash

# Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "BUILDING QED EVENT FROM GO.MOD"
msg="message"
version="v1.0.0"
hash=$(echo -n $version | sha256sum | cut -d' ' -f1)
salt=$(echo -n $(hostname) | sha256sum | cut -d' ' -f1)

echo "
{
	\"msg\": \"$salt $msg\",
	\"version\": \"$version\",
	\"hash\": \"$hash\"
}
" > event1.json
echo -e "\t RESULTING QED EVENT:"
cat event1.json
read -p "Press intro to continue"

echo -e "\t ASKING FOR MEMBERSHIP PROOF:"
go run ../main.go client membership --api-key key --insecure --event "$(cat event1.json)" --log info --version 10
