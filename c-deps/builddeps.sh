#!/usr/bin/env bash

BASE=$(pwd)
LIBS="$BASE/libs"

# build jemalloc
if [ ! -f $LIBS/libjemalloc.so.2 ]; then 
	cd jemalloc
	bash autogen.sh
	make -j8
	cp lib/libjemalloc.so.2 ../libs/
	cp lib/libjemalloc.a ../libs
	cd ../libs
	ln -s ibjemalloc.so.2 libjemalloc.so
fi

cd $BASE

# # build bzip2 shared lib
# if [ ! -f $LIBS/libbz2.so.1.0.6 ]; then 
# 	cd bzip2
# 	make -j8 -f  Makefile-libbz2_so
# 	make -j8
# 	cp libbz2.so.1.0.6 ../libs/
# 	cp libbz2.a ../libs/
# 	cd ../libs/
# 	ln -s libbz2.so.1.0.6 libbz2.so.1
# 	ln -s libbz2.so.1.0.6 libbz2.so
# fi

# cd $BASE

# # build lz4 shared lib
# if [ ! -f $LIBS/liblz4.so.1.8.3 ]; then 
# 	cd lz4
# 	make -j8
# 	cp lib/liblz4.so.1.8.3 ../libs/
# 	cp lib/liblz4.a ../libs
# 	cd ../libs
# 	ln -s liblz4.so.1.8.3 liblz4.so.1
# 	ln -s liblz4.so.1.8.3 liblz4.so
# fi

# cd $BASE

# build snappy shared lib
if [ ! -f $LIBS/libsnappy.so.1.1.7 ]; then 
	cd snappy
	mkdir build
	cd build
	cmake ../
	sed -i.bak  s/BUILD_SHARED_LIBS:BOOL=OFF/BUILD_SHARED_LIBS:BOOL=ON/g CMakeCache.txt
	make -j8
	cp libsnappy.so.1.1.7 ../../libs/
	cp libsnappy.a ../../libs/
	cp snappy-stub-public.h ../
	cd ../../libs/
	ln -s libsnappy.so.1.1.7 libsnappy.so.1
	ln -s libsnappy.so.1.1.7 libsnappy.so
fi

cd $BASE

# # build zlib shared lib
# if [ ! -f $LIBS/libz.so.1.2.9 ]; then 
# 	cd zlib
# 	./configure
# 	make -j8
# 	cp libz.so.1.2.9 ../libs/
# 	cp libz.a ../libs/
# 	cd ../libs
# 	ln -s libz.so.1.2.9 libz.so.1
# 	ln -s libz.so.1.2.9 libz.so
# fi

# cd $BASE

# # build zstd shared lib
# if [ ! -f $LIBS/libzstd.so.0.4.2 ]; then 
# 	cd zstd
# 	cd lib
# 	make -j8
# 	cp libzstd.so.0.4.2 ../../libs
# 	cp libzstd.a ../../libs
# 	cd ../../libs
# 	ln -s libzstd.so.0.4.2 libzstd.so.0
# 	ln -s libzstd.so.0.4.2 libzstd.so
# fi

cd $BASE

# build rocksdb shared with those libraries 
cd rocksdb
mkdir -p build
cd build


	# -DWITH_LZ4=ON -DLZ4_LIBRARIES="$LIBS/liblz4.a" -DLZ4_INCLUDE_DIR="$BASE/lz4/lib/" \
	# -DWITH_ZSTD=ON -DZSTD_LIBRARIES="$LIBS/libzstd.a" -DZSTD_INCLUDE_DIR="$BASE/zstd" \
	# -DWITH_BZ2=ON -DBZIP2_LIBRARIES="$LIBS/libbz2.a" -DBZIP2_INCLUDE_DIR="$BASE/bzip2" \
	# -DWITH_ZLIB=ON -DZLIB_LIBRARIES="$LIBS/libz.a" -DZLIB_INCLUDE_DIR="$BASE/zlib" \

cmake -DWITH_GFLAGS=OFF  -DPORTABLE=ON \
	-DWITH_SNAPPY=ON -DSNAPPY_LIBRARIES="$LIBS/libsnappy.a" -DSNAPPY_INCLUDE_DIR="$BASE/snappy" \
	-DWITH_JEMALLOC=ON -DJEMALLOC_LIBRARIES="$LIBS/libjemalloc.a" -DJEMALLOC_INCLUDE_DIR="$BASE/jemalloc/include" \
	-DCMAKE_BUILD_TYPE=Release -DUSE_RTTI=1 ../
make -j8 rocksdb

cp librocksdb.a* ../libs
