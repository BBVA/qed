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

if [ -z "$1" -o  -z "$2" -o  -z "$3" ]; then
	echo usage:
	echo $0 pkg_name pkg_version pkg_url
	exit -1
fi
_echo() {(>&2 echo "$1")}

dir=$(mktemp -d)
wget -q -P $dir $3 
pkg_digest=$(sha256sum $dir/* | cut -f 1 -d ' ')

# Entry data structure
read -r -d '' entry <<EOF
{
	"pkg_name": "$1",
	"pkg_version": "$2",
	"pkg_digest": "$pkg_digest"
}
EOF

_echo "Package"
_echo "$entry" 
_echo "saved in $dir"

echo "$entry"

