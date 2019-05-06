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
msg=$(cat build/project/go.mod | xargs )
version="v1.3.0"
hash=$(sha256sum build/project/go.mod | cut -d' ' -f1)
salt=$(echo -n $(hostname) | sha256sum | tr -d ' ' | tr -d '-')

echo "
{
	\"msg\": \"$salt $msg\",
	\"version\": \"$version\",
	\"hash\": \"$hash\"
}
" > event0.json
echo -e "\t RESULTING QED EVENT:"
cat event0.json
read -p "Press intro to continue"

echo -e "\t ASKING FOR MEMBERSHIP PROOF:"
echo 'go run ../main.go client membership --api-key key --insecure --event "$(cat event0.json)" --log info'
