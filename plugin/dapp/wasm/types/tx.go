// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

type CreateOrCallWasmContract struct {
	Value CreateOrCallWasmContractValue
}

type CreateOrCallWasmContractValue interface {
	WasmContractActionType() int
}

// 创建wasm合约
type CreateWasmContractPara struct {
	// 合约字节码
	Code []byte `json:"code"`
	// abi字符串
	Abi string `json:"abi"`
	// 合约别名，方便识别
	Name string `json:"name"`
	// 交易备注
	Note string `json:"note"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}

func (CreateWasmContractPara) WasmContractActionType() int {
	return CreateWasmContractAction
}

// 调用wasm合约
type CallWasmContractPara struct {
	Name string `json:"name"`
	// 交易备注
	Note string `json:"note"`
	// 执行动作名称
	ActionName string `json:"actionName"`
	// 执行参数,abi格式 数据格式
	ActionData []byte `json:"actionData"`
	// Fee 交易手续费
	Fee int64 `json:"fee"`
}

func (CallWasmContractPara) WasmContractActionType() int {
	return CallWasmContractAction
}
