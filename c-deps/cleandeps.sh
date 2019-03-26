#!/usr/bin/env bash

for i in jemalloc snappy rocksdb
do
    make -C $i clean
done