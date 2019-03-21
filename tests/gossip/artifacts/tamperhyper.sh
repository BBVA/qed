#!/bin/bash

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

if [ -z "$1" -o -z "$2" ]; then
	echo usage:
	echo $0 entry_file value_version
	exit -1
fi
event=$(cat $1 | sed 's/\n//g' |base64 -w0)

payload=$(cat <<-EOF
	{"event":"$event" , "value": $2}
EOF
)
curl -X POST  http://localhost:18800/hyper/add -H 'Api-Key: foo' -d "$payload"
