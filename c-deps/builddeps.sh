#!/usr/bin/env bash

set -e

BASE=$(pwd)
LIBS="$BASE/libs"
mkdir -p $LIBS

# build jemalloc
if [ ! -f $LIBS/libjemalloc.a ]; then
	cd jemalloc
	bash autogen.sh
	make -j8
	cp lib/libjemalloc.a ../libs
	cd ../libs
fi

cd $BASE

# build snappy shared lib
if [ ! -f $LIBS/libsnappy.a ]; then
	cd snappy
	mkdir -p build
	cd build
	cmake ../
	# sed -i.bak  s/BUILD_SHARED_LIBS:BOOL=OFF/BUILD_SHARED_LIBS:BOOL=ON/g CMakeCache.txt
	make -j8

	cp libsnappy.a ../../libs/
	cp snappy-stubs-public.h ../
	cd ../../libs/
fi

cd $BASE

if [ ! -f $LIBS/librocksdb.a ]; then
	# build rocksdb shared with those libraries
	cd rocksdb
	mkdir -p build
	cd build
	cmake -DWITH_GFLAGS=OFF -DPORTABLE=ON \
	-DWITH_SNAPPY=ON -DSNAPPY_LIBRARIES="$LIBS/libsnappy.a" -DSNAPPY_INCLUDE_DIR="$BASE/snappy" \
	-DWITH_JEMALLOC=ON -DJEMALLOC_LIBRARIES="$LIBS/libjemalloc.a" -DJEMALLOC_INCLUDE_DIR="$BASE/jemalloc/include" \
	-DCMAKE_BUILD_TYPE=Release -DUSE_RTTI=1 ../
	make -j8 rocksdb

	cp librocksdb.a ../../libs
fi
