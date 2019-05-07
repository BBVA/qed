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

export GO111MODULE=on

mkdir -p build
cd build
../get_project.sh
cd ..

echo -n'' | ./membership_event0.sh > /tmp/membership_result
membership_check=$(grep "true" /tmp/membership_result)

if [[ "$membership_check" = " Exists: true" ]];
then
    ./membership_event0.sh
else
    echo -e "EVENT NOT FOUND\nBUILD FAILED!"
    exit 1
fi
read -p "Press intro to continue"

echo -e "\n GETTING SNAPSHOT INFO. FROM SNAPSHOT STORE \n"
./get_snapshot.sh 0
read -p "Press intro to continue"

echo -e "\n VERIFY DEPENDENCIES AUTHENTICITY\n"
./membership_verify_event0.sh
read -p "Press intro to continue"

if [[ "$?" -eq 0 ]]
then
    echo "Building project"
    cd build/project
    go build -o gin
    echo "Generating artifact in build/project"
    sleep 1
    echo "gin binary file created"
else
    echo "Verification failed. The project has been tampered!"
fi

