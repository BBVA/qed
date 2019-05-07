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
echo "BUILDING QED EVENT"
cd project
commit_hash=$(git rev-parse HEAD)
src_hash=$(echo $(find . -type f -not -path "./.git/*" -exec sha256sum {} \; | sort -k2) | sha256sum | cut -d' ' -f1)
cd ../
echo "
{
	\"commit_hash\": \"$commit_hash\",
	\"src_hash\": \"$src_hash\"
}
" > event0.json
echo -e "\t RESULTING QED EVENT:"
cat event0.json
/tmp/qed client --api-key key --insecure add --event "$(cat event0.json)" --log info
