#!/bin/bash

function _readlink() { (
  # INFO: readlink does not exist on OSX ¯\_(ツ)_/¯
  cd $(dirname $1)
  echo $PWD/$(basename $1)
) }

# Deployment options
CGO_LDFLAGS_ALLOW='.*'
QED="go run $GOPATH/src/github.com/bbva/qed/main.go"

pub=$(_readlink ./config_files)
tdir=$(mktemp -d /tmp/qed_build.XXX)

sign_path=${pub}
cert_path=${pub}

(
cd ${tdir}

if [ ! -f ${node_path} ]; then (
    version=0.17.0
    folder=node_exporter-${version}.linux-amd64
    link=https://github.com/prometheus/node_exporter/releases/download/v${version}/${folder}.tar.gz
    wget -qO- ${link} | tar xvz -C ./
    cp ${folder}/node_exporter ${node_path}
) fi

if [ ! -f ${sign_path} ]; then
    #build shared signing key
    $QED generate signerkeys --path ${sign_path}
fi

if [ ! -f ${sign_path} ]; then
    #build shared signing key
    $QED generate tlscerts --path ${cert_path} --host 127.0.0.1
fi

)

export GOOS=linux
export GOARCH=amd64
#build server binary
go build -o ${pub}/qed ../../
