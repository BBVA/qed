#!/usr/bin/env bash

#  Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

GO111MODULE=on
CGO_LDFLAGS_ALLOW='.*'

Tag=$(git tag --points-at HEAD | grep v[0-9.0-9.0-9])
FullCommit=$(git rev-parse HEAD)

if [[ $(uname) == "Darwin" ]]
then
    Date=$(date -u +'%Y/%m/%dT%XUTC')
else
    Date=$(date --utc +'%Y/%m/%dT%XUTC')
fi

mkdir -p dist

echo "Building binary"
go build -ldflags="-s -w -X github.com/bbva/qed/build.tag=${Tag} -X github.com/bbva/qed/build.rev=${FullCommit} -X github.com/bbva/qed/build.utcTime=${Date}" -o dist/qed

cp README.rst dist/README.rst
cp LICENSE dist/LICENSE

tar -C dist -zcvf qed_${Tag}_linux_amd64.tar.gz .
mv qed_${Tag}_linux_amd64.tar.gz dist
md5sum dist/qed_${Tag}_linux_amd64.tar.gz > dist/qed_${Tag}_checksum.txt

if [ "$?" != 0 ]
then
    echo "Error: QED build failed"
    exit 1
fi

if [ ! -z "${DOCKER_BUILD}" ]
then
    echo "Building Docker image"
    docker build -t bbvalabs/qed:${Tag} .

    if [ "$?" != 0 ]
    then
        echo "Error: QED Docker build failed"
        exit 1
    fi
else
    echo "Skipping Docker Build"
fi
