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

FROM golang:1.12.1

ENV GO111MODULE=on
ENV CGO_LDFLAGS_ALLOW='.*'

WORKDIR /go/src/github.com/bbva/qed

# Install deps.
RUN apt update -qq && apt install -qq -y autoconf cmake

# Build C deps. 
# This step acts as cache to avoid recompiling when Go code changes.
RUN git clone https://github.com/BBVA/qed.git .         &&\
    git submodule update --init --recursive             &&\
    cd c-deps                                           &&\
    ./builddeps.sh                                      &&\
    go mod download

# Build QED, Storage binary and riot
RUN go build -o /usr/local/bin/qed                                   &&\
    go build -o /usr/local/bin/riot testutils/riot/riot.go           &&\
    go build -o /usr/local/bin/storage testutils/notifierstore.go

# Clean
RUN rm -rf /var/lib/apt/lists/*

FROM ubuntu:19.10

COPY --from=0 /usr/local/bin/qed /usr/local/bin/qed
COPY --from=0 /usr/local/bin/riot /usr/local/bin/riot
COPY --from=0 /usr/local/bin/storage /usr/local/bin/storage

RUN /usr/local/bin/qed generate signerkeys
