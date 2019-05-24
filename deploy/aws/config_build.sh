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
cert_path=${pub}/server.crt
key_path=${pub}/server.key
node_path=${pub}/node_exporter

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
    $QED generate keypair --path ${sign_path}
fi

if [ ! -f ${cert_path} ] && [ ! -f ${key_path} ]; then

    #build shared server cert
    openssl req \
        -newkey rsa:2048 \
        -nodes \
        -days 3650 \
        -x509 \
        -keyout ca.key \
        -out ca.crt \
        -subj "/CN=*"
    openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout server.key \
        -out server.csr \
        -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=*"
    openssl x509 \
        -req \
        -days 365 \
        -sha256 \
        -in server.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -CAcreateserial \
        -out server.crt \
        -extfile <(echo subjectAltName = IP:127.0.0.1)

    cp server.crt ${cert_path}
    cp server.key ${key_path}

fi

)

export GOOS=linux
export GOARCH=amd64
#build server binary
go build -o ${pub}/qed ../../
