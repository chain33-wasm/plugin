# wasmbuilder_chain33
wasm contract run platform builder for chain33 based on ubuntu:16.04

#create the image 
docker build . -t wasmbuilder_chain33:0.9

#startup a docker container to build chain33 where wasm contract can be executed 
docker run -it --name docker_build_wasm --rm -v $GOPATH/src/github.com/33cn/plugin/:/go/src/github.com/33cn/plugin/ -v /usr/local/go/:/usr/local/go/ wasmbuilder_chain33:0.9 /bin/bash
