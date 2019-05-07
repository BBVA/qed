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

echo "RELEASE STAGE"
echo "CHECKING ARTIFACT"
read -p "Press intro to continue"
echo "ARTIFACT APPROVED. GENERATING QED EVENT..."
./add_event2.sh
echo "UPLOAD TO ARTIFACTS REPOSITORY.."
read -p "Press intro to continue"
echo "DONE"

