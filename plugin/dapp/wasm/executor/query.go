package executor

//#cgo CFLAGS: -Iexternal/abi/include/
//#cgo LDFLAGS: -Lexternal/abi/lib/ -labiconv -lboost_date_time -lstdc++
//#include <stdio.h>
//#include <stdlib.h>
//#include "abieos.h"
import "C"

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/common/address"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"unsafe"
)


// Query_CheckContractNameExist 确认是否存在该wasm合约，
func (wasm *WASMExecutor) Query_CheckContractNameExist(in *wasmtypes.CheckWASMContractNameReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext()
	return wasm.checkContractNameExists(in)
}

//TODO:稍后再支持
// Query_EstimateGasFunc 试运行wasm合约用来估算gas的消耗量
//func (wasm *WASMExecutor) Query_EstimateGas(in *wasmtypes.EstimateWASMGasReq) (types.Message, error) {
//	if in == nil {
//		return nil, types.ErrInvalidParam
//	}
//	return wasm.estimateGas(in)
//}

//TODO:稍后再支持
// Query_DebugCode 调试wasm代码
//func (wasm *WASMExecutor) Query_DebugCode(in *wasmtypes.EstimateWASMGasReq) (types.Message, error) {
//	if in == nil {
//		return nil, types.ErrInvalidParam
//	}
//	return wasm.debugCode(in)
//}

// Query_DebugCode 调试wasm代码


func (wasm *WASMExecutor) Query_ConvertJson2Abi(in *wasmtypes.ConvertJson2AbiReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext()

	resp := &wasmtypes.Json2AbiResponse{}
	contractAddr := address.ExecAddress(types.ExecName(in.ContractName))
	abi := wasm.mStateDB.GetAbi(contractAddr)

	resp.AbiData = genAbiData(string(abi), in.ContractName, in.ActionName, in.AbiDataInJson)

	return resp, nil
}

func (wasm *WASMExecutor) Query_WasmGetAbi(in *types.ReqAddr) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext()

	resp := &wasmtypes.WasmGetAbiResp{}
	resp.Abi = wasm.mStateDB.GetAbi(in.Addr)
	return resp, nil
}

// Query_WasmGetContractTable 查询某个wasm合约的指定表的结果
func (wasm *WASMExecutor) Query_WasmGetContractTable(in *wasmtypes.WasmQueryContractTableReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext()

	return wasm.getContractTable(in)
}

func genAbiData(contractAbi, contractName, actionName, abiJson string) []byte {
	contract := C.CString(contractName)
	defer C.free(unsafe.Pointer(contract))

	action := C.CString(actionName)
	defer C.free(unsafe.Pointer(action))

	abii := C.CString(contractAbi)
	defer C.free(unsafe.Pointer(abii))

	para := C.CString(abiJson)
	//para := C.CString("{\"user\":\"abcdf\"}")
	defer C.free(unsafe.Pointer(para))

	var abisize C.int
	abidata := C.genAbiFromJson(contract, action, abii, para, &abisize)
	defer C.free(unsafe.Pointer(abidata))

	abislice := C.GoBytes(unsafe.Pointer(abidata), abisize)
	return abislice
}
