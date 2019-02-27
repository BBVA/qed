#!/usr/bin/env bash

cd rocksdb
DEBUG_LEVEL=0
PORTABLE=1

for lib in libbz2.a liblz4.a libsnappy.a libz.a libzstd.a; do
	make $lib
done
make static_lib
strip *.a
cd ..


