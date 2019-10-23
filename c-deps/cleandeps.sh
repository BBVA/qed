#!/usr/bin/env bash

. ./vars

for i in jemalloc snappy rocksdb
do
    ${MAKE_CMD} -C $i clean
done
