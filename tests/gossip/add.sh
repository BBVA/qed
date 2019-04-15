#!/bin/bash

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

if ! which envsubst
then
    echo -e "Please install envsubst. OSX -> brew install gettext ; brew link --force gettext"
    exit 1
fi

# client options
CLIENT_CONFIG=()
CLIENT_CONFIG+=("--log debug")
CLIENT_CONFIG+=("--endpoints http://127.0.0.1:8800")
config=$(echo ${CLIENT_CONFIG[@]} | i=0 envsubst )

go run $GOPATH/src/github.com/bbva/qed/main.go client add $config --event $1
