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

mkdir -p deploy
echo "DOWNLOADING ARTIFACT"
cp build/project/gin deploy/
read -p "Press intro to continue"
./membership_event2.sh

echo -e "\n GETTING SNAPSHOT INFO. FROM SNAPSHOT STORE \n"
./get_snapshot.sh 2
read -p "Press intro to continue"

echo "VERIFY ARTIFACT.."
./membership_verify_event2.sh

read -p "Press intro to continue"
echo -e "\n GETTING SNAPSHOT INFO FOR VERSION 0. FROM SNAPSHOT STORE \n"
./get_snapshot.sh 0
read -p "Press intro to continue"
echo -e "\n GETTING SNAPSHOT INFO FOR VERSION 2. FROM SNAPSHOT STORE \n"
./get_snapshot.sh 2
read -p "Press intro to continue"

echo -e "\n EXECUTING INCREMENTAL PROOF \n"
./incremental_start0_end2.sh

read -p "Press intro to continue"
echo "DEPLOYING ARTIFACTS.."
read -p "Press intro to continue"
echo "DONE"

