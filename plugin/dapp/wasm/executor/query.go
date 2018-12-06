package executor

import (
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
)


// Query_CheckAddrExistsWasm 确认是否存在该wasm合约地址，
func (wasm *WASMExecutor) Query_CheckAddrExistsWasm(in *wasmtypes.CheckWASMAddrReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return wasm.checkAddrExists(in)
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
func (wasm *WASMExecutor) Query_WasmGetAbi(in *types.ReqAddr) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	resp := &wasmtypes.WasmGetAbiResp{}
	resp.Abi = wasm.mStateDB.GetAbi(in.Addr)
	return resp, nil
}

// Query_WasmGetContractTable 查询某个wasm合约的指定表的结果
func (wasm *WASMExecutor) Query_WasmGetContractTable(in *wasmtypes.WasmQuery) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	return wasm.getContractTable(in)
}