package executor

//#cgo CFLAGS: -Iexternal/abi/include/
//#cgo LDFLAGS: -Lexternal/abi/lib/ -labiconv -lboost_date_time -lstdc++
//#include <stdio.h>
//#include <stdlib.h>
//#include "abieos.h"
import "C"

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/golang/protobuf/proto"
	"os"
	"regexp"
	"unsafe"
	"encoding/hex"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
)


// Query_CheckContractNameExist 确认是否存在该wasm合约，
func (wasm *WASMExecutor) Query_CheckContractNameExist(in *wasmtypes.CheckWASMContractNameReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))
	return wasm.checkContractNameExists(in)
}


// Query_EstimateGasCreateContract 试运行wasm合约用来估算gas的消耗量
func (wasm *WASMExecutor) Query_EstimateGasCreateContract(in *wasmtypes.EstimateCreateContractReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	createDataGas := (uint64(len(in.Code)) + uint64(len(in.Abi))) * loccom.CreateDataGas
	resp := &wasmtypes.EstimateWASMGasResp{}
	resp.Gas = createDataGas
	return resp, nil
}

// Query_EstimateGasCallContract 试运行wasm合约用来估算gas的消耗量
func (wasm *WASMExecutor) Query_EstimateGasCallContract(in *wasmtypes.EstimateCallContractReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))

	//resp := &wasmtypes.Json2AbiResponse{}
	contractAddr := address.ExecAddress(in.Execer)
	abi := wasm.mStateDB.GetAbi(contractAddr)

	AbiData := genAbiData(string(abi), in.Execer, in.ActionName, string(in.ActionData))
	in.ActionData = AbiData
	action := &wasmtypes.WasmContractAction{
		Value: &wasmtypes.WasmContractAction_CallWasmContract{
			CallWasmContract: &wasmtypes.CallWasmContract{
				GasLimit:   uint64(in.GasLimit),
				GasPrice:   1,
				Note:       "",
				VmType:     wasmtypes.VMBinaryen, //当前只支持binaryen解释执行的方式
				ActionName: in.ActionName,
				ActionData: AbiData,
			},
		},
		Ty: wasmtypes.CallWasmContractAction,
	}
	tx, err := createRawWasmTx(action, in.Execer, int64(in.GasLimit))
	if err != nil {
		return nil, err
	}

	useGas , err := wasm.estimateGasCall(in, tx)
	if err != nil {
		return nil, err
	}
	resp := &wasmtypes.EstimateWASMGasResp{}
	resp.Gas = useGas
	return resp, nil
}

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
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))

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
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))

	resp := &wasmtypes.WasmGetAbiResp{}
	resp.Abi = wasm.mStateDB.GetAbi(in.Addr)
	return resp, nil
}

// Query_WasmFuzzyGetContractTable 模糊查询某个wasm合约的指定表的结果
func (wasm *WASMExecutor) Query_WasmFuzzyGetContractTable(in *wasmtypes.WasmFuzzyQueryTableReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	contractName := in.ContractName
	if !bytes.Contains([]byte(in.ContractName), []byte(wasmtypes.UserWasmX)) {
		contractName = wasmtypes.UserWasmX + contractName
	}

	incp := *in
	incp.ContractName = contractName

	return wasm.fuzzyGetContractTable(&incp)
}

// Query_WasmGetContractTable 查询某个wasm合约的指定表的结果
func (wasm *WASMExecutor) Query_WasmGetContractTable(in *wasmtypes.WasmQueryContractTableReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	contractName := in.ContractName
	if !bytes.Contains([]byte(in.ContractName), []byte(wasmtypes.UserWasmX)) {
		contractName = wasmtypes.UserWasmX + contractName
	}

	incp := *in
	incp.ContractName = contractName

	return wasm.getContractTable(&incp)
}

// Query_CreateTx 创建交易对象
func (wasm *WASMExecutor) Query_CreateWasmContract(in *wasmtypes.CreateContrantReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	execer := types.GetRealExecName([]byte(in.Name))
	if bytes.HasPrefix(execer, []byte(wasmtypes.UserWasmX)) {
		execer = execer[len(wasmtypes.UserWasmX):]
	}

	execerStr := string(execer)
	nameReg, err := regexp.Compile(wasmtypes.NameRegExp)
	if !nameReg.MatchString(execerStr) {
		fmt.Fprintln(os.Stderr, err, "Wrong wasm contract name format, which should be a-z and 0-9 ")
		return nil, wasmtypes.ErrWrongContracName
	}

	if len(execerStr) > 16 || len(execerStr) < 4 {
		fmt.Fprintln(os.Stderr, "wasm contract name's length should be within range [4-16]")
		return nil, wasmtypes.ErrWrongContracNameLen
	}
	action := &wasmtypes.WasmContractAction{
		Value: &wasmtypes.WasmContractAction_CreateWasmContract{
			CreateWasmContract: &wasmtypes.CreateWasmContract{
				GasLimit: uint64(in.Fee),
				GasPrice: 1,
				Code:     in.Code,
				Abi:      in.Abi,
				Name:     types.ExecName(wasmtypes.UserWasmX + execerStr),
				Note:     in.Note,
			},
		},
		Ty: wasmtypes.CreateWasmContractAction,
	}
	createRsp, err := createRawWasmTx(action, wasmtypes.WasmX, in.Fee)
	result := hex.EncodeToString(types.Encode(createRsp))
	relpydata := &types.ReplyString{Data:result}
	return relpydata, err

}

// Query_CreateTx 创建交易对象
func (wasm *WASMExecutor) Query_CallWasmContract(in *wasmtypes.CallContractReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))

	contractName := in.Name
	if !bytes.Contains([]byte(contractName), []byte(wasmtypes.UserWasmX)) {
		contractName = wasmtypes.UserWasmX + contractName
	}

	//resp := &wasmtypes.Json2AbiResponse{}
	contractAddr := address.ExecAddress(types.ExecName(contractName))
	abi := wasm.mStateDB.GetAbi(contractAddr)

	AbiData := genAbiData(string(abi), contractName, in.ActionName, in.DataInJson)

	action := &wasmtypes.WasmContractAction{
		Value: &wasmtypes.WasmContractAction_CallWasmContract{
			CallWasmContract: &wasmtypes.CallWasmContract{
				GasLimit:   uint64(in.Fee),
				GasPrice:   1,
				Note:       in.Note,
				VmType:     wasmtypes.VMBinaryen, //当前只支持binaryen解释执行的方式
				ActionName: in.ActionName,
				ActionData: AbiData,
			},
		},
		Ty: wasmtypes.CallWasmContractAction,
	}
	createRsp, err := createRawWasmTx(action, contractName, in.Fee)
	result := hex.EncodeToString(types.Encode(createRsp))
	replydata:= &types.ReplyString{Data:result}
	return replydata, err
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

func createRawWasmTx(action proto.Message, wasmName string, fee int64) (*types.Transaction, error) {
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(wasmName)),
		Payload: types.Encode(action),
		To:      address.ExecAddress(types.ExecName(wasmName)),
	}
	tx, err := types.FormatTx(string(tx.Execer), tx)
	if err != nil {
		return nil, err
	}
	if tx.Fee < fee {
		tx.Fee = fee
	}
	return tx, nil
}
