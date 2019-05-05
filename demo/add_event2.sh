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

name="gin"
version="v1.3.0"
hash=$(sha256sum release/gin | sed 's/  release\/gin//g' )
salt=$(echo -n $(hostname) | sha256sum | tr -d ' ' | tr -d '-')
msg="$version $name"

echo "
{
	\"msg\": \"$salt $msg $version\",
	\"version\": \"$version\",
	\"hash\": \"$hash\"
}
" > event2.json
go run ../main.go client --api-key key --insecure add --event "$(cat event2.json)" --log info
