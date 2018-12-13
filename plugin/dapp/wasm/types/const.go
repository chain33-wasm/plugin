// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

const (
	// EVM本执行器支持的查询方法
	CheckNameExistsFunc    = "CheckContractNameExist"
	EstimateGasWasm        = "EstimateGasWasm"
	WasmDebug              = "WasmDebug"
	WasmGetAbi             = "WasmGetAbi"
	ConvertJson2Abi        = "ConvertJson2Abi"
	WasmGetContractTable   = "WasmGetContractTable"
	GAS_EXHAUSTED_ERR_CODE = 0x81234567
	WasmX                  = "wasm"
	UserWasmX              = "user.wasm."
	CreateWasmContractStr  = "CreateWasmContract"
	CallWasmContractStr    = "CallWasmContract"
	//NameRegExp             = "[a-z0-9]"^[a-z]+\[[0-9]+\]$
	NameRegExp             = "^[a-z0-9]+$"
	AccountOpFail          = int(-1)
	AccountOpSuccess       = int(0)
    RetryNum               = int(10)
	GRPCRecSize            = 5 * 30 * 1024 * 1024
)

//wasm contract action
const (
	CreateWasmContractAction = 1 + iota
	CallWasmContractAction
)

const (
	// log for wasm
	// 合约代码日志
	TyLogContractDataWasm = iota + 100
	// 合约状态数据日志
	TyLogContractStateWasm
	// 合约调用日志
	TyLogCallContractWasm
	//
	TyLogStateChangeItemWasm
	//
	TyLogCreateUserWasmContract

	//用于wasm合约输出可读信息的日志记录，尤其是query的相关信息
	//为什么不将该种信息类型的获取不放置在query中呢，因为query的操作
	// 中是不含交易费的，如果碰到恶意的wasm合约，输出无限长度的信息，
	// 会对我们的wasm合约系统的安全性造成威胁，基于这样的考虑我们
	TyLogOutputItemWasm
)

const (
	VMWavm = iota
	VMBinaryen
)
