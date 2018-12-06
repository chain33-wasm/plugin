package common

import (
	"fmt"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types/executor/wasm"

)

type WasmContratOpType int

const (

	// 本执行器前缀
	WasmPrefix = types.UserWasmX
	// 本执行器名称
	ExecutorName = types.WasmX

	CallWasmContrcat   = WasmContratOpType(1)
	CreateWasmContrcat = WasmContratOpType(2)


	// EVM本执行器支持的查询方法
	CheckAddrExistsFunc = "CheckAddrExistsWasm"
	EstimateGasFunc     = "EstimateGasWasm"
	WasmDebug           = "WasmDebug"
	WasmGetAbi          = "WasmGetAbi"
	//QueryFromContract   = "QueryFromContract"
	QueryFromContract   = wasm.QueryFromContract

	//各种数据存储前缀
	WasmContractCodePrefix     = "mavl-wasm-contract-code"
	WasmContractABIPrefix      = "mavl-wasm-contract-abi"
	WasmContractKvPrefix       = "mavl-wasm-contract-kv"
	WasmContractAliasPrefix    = "local-wasm-alias"
)

func CalcWasmContractName(txhash []byte) string {
	return types.ExecName(WasmPrefix) + common.ToHex(txhash)
}

//mavl-wasm-contract-code-user.wasm.contrAddr --->>wasm byte opcode
func CalcWasmContractCodeKey(contractAddr string) (key []byte) {
	return []byte(fmt.Sprintf(WasmContractCodePrefix + "-%s", WasmPrefix+contractAddr))
}

//mavl-wasm-contract-abi-user.wasm.contrAddr --->>wasm abi byte
func CalcWasmContractABIKey(contractAddr string) (key []byte) {
	return []byte(fmt.Sprintf(WasmContractABIPrefix + "-%s", WasmPrefix+contractAddr))
}

//通过合约别名获取真正的合约地址，进而获取合约代码
//local-wasm-alias-contractAlias--->>contrAddr
func calcWasmContrAliasKey(alias string) []byte {
	return []byte(fmt.Sprintf(WasmContractAliasPrefix+"-%s", WasmPrefix+alias))
}
