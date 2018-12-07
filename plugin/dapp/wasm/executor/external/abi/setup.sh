#!/bin/bash

if [ ! -d "build" ]; then
  mkdir build
fi

rm build/* -rf
cd build
cmake .. && make
cd ..

if [ ! -d "lib" ]; then
  mkdir lib
fi
rm lib/* -fr

if [ ! -d "include" ]; then
  mkdir include
fi

cp build/libabiconv.a lib/
cp src/abieos.h include/

