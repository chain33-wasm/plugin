#!/bin/bash
set -x

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "current dir is:"${CURRENT_DIR}

bash ${CURRENT_DIR}/genMakeFile.sh && cd ${CURRENT_DIR}/build && make -j4


cp adapter/libwasm_adapter.a ../lib
cp wasm/binaryen/lib/libasmjs.a ../lib
cp wasm/binaryen/lib/libast.a ../lib
cp wasm/binaryen/lib/libwasm.a ../lib
cp wasm-jit/Source/WASM/libWASM.a ../lib
cp wasm-jit/Source/WAST/libWAST.a ../lib
cp fc/libfc.a ../lib
cp serialize/libwasm_serialize.a ../lib


