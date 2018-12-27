# wasmbuilder_chain33
wasm contract run platform builder for chain33 based on ubuntu:16.04

#create the image 
docker build . -t wasmbuilder_chain33:0.9

#startup a docker container to build chain33 where wasm contract can be executed 
docker run -it --name build_wasm --rm -v $GOPATH/src/github.com/33cn/plugin/:/go/src/github.com/33cn/plugin/ wasmbuilder_chain33:0.9 /bin/bash

docker run -it --name build_wasm --rm -v ~/work/go/src/github.com/33cn/plugin/:/go/src/github.com/33cn/plugin/ chain33wasm_build:1.0 /bin/bash
