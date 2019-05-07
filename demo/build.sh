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

echo "BUILD STAGE"
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

echo -e "\n GETTING SNAPSHOT INFO FOR VERSION 0. FROM SNAPSHOT STORE \n"
./get_snapshot.sh 0
read -p "Press intro to continue"

echo -e "\n VERIFY DEPENDENCIES AUTHENTICITY\n"
./membership_verify_event0.sh
read -p "Press intro to continue"

if [[ "$?" -eq 0 ]]
then
    echo "BUILDING PROJECT"
    cd build/project
    go build -o /tmp/gin
    echo "GENERATING ARTIFACT IN BUILD/PROJECT"
    sleep 1
    echo "GIN BINARY CREATED IN /tmp/gin"
    cd ../..
else
    echo "VERIFICATION FAILED. THE PROJECT HAS BEEN TAMPERED!"
fi

echo -n'' | ./membership_event1.sh > /tmp/membership_result
membership_check=$(grep "true" /tmp/membership_result)

if [[ "$membership_check" = " Exists: true" ]];
then
    echo "EVENT WITH VERSION 1 ALREADY ADDED"
else
    echo "GENERATING INTERMEDIATE EVENT"
    ./add_event1.sh
fi

rm -f archived/gin
if [ ! -f archived/gin ];
then
    mkdir -p archived
    cp /tmp/gin archived/gin
fi

echo -n'' | ./membership_event2.sh > /tmp/membership_result
membership_check=$(grep "true" /tmp/membership_result)

if [[ "$membership_check" = " Exists: true" ]];
then
    echo "ARTIFACT ALREADY RELEASED"
else
    ./release.sh
fi



