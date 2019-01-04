// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"strings"
	"reflect"
)

var (
	elog = log.New("module", "exectype.wasm")

	actionName = map[string]int32{
		CreateWasmContractStr: CreateWasmContractAction,
		CallWasmContractStr:   CallWasmContractAction,
	}
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(WasmX))
	// init executor type
	types.RegistorExecutor(WasmX, NewType())
}

// WasmType EVM类型定义
type WasmType struct {
	types.ExecTypeBase
}

// NewType 新建EVM类型对象
func NewType() *WasmType {
	c := &WasmType{}
	c.SetChild(c)
	return c
}

// GetPayload 获取消息负载结构
func (wasm *WasmType) GetPayload() types.Message {
	return &WasmContractAction{}
}

// ActionName 获取ActionName
func (wasm WasmType) ActionName(tx *types.Transaction) string {
	// 这个需要通过合约交易目标地址来判断Action
	// 如果目标地址为空，或为wasm的固定合约地址，则为创建合约，否则为调用合约
	if strings.EqualFold(tx.To, address.ExecAddress(types.ExecName(WasmX))) {
		return "createWasmContract"
	}
	return "callWasmContract"
}

// GetTypeMap 获取类型映射
func (wasm *WasmType) GetTypeMap() map[string]int32 {
	return actionName
}

// GetRealToAddr 获取实际地址
func (wasm WasmType) GetRealToAddr(tx *types.Transaction) string {
	var action WasmContractAction
	err := types.Decode(tx.Payload, &action)
	if err == nil {
		return tx.To
	}

	return ""
}

// Amount 获取金额
func (wasm WasmType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}


// GetLogMap 获取日志类型映射
func (wasm *WasmType) GetLogMap() map[int64]*types.LogInfo {
	logInfo := map[int64]*types.LogInfo{
		TyLogContractDataWasm:       {Ty: reflect.TypeOf(LogWASMContractData{}), Name: "LogContractDataWasm"},
		TyLogContractStateWasm:       {Ty: reflect.TypeOf(WASMContractState{}), Name: "LogContractStateWasm"},
		TyLogCallContractWasm:      {Ty: reflect.TypeOf(ReceiptWASMContract{}), Name: "LogCallContractWasm"},
		TyLogStateChangeItemWasm: {Ty: reflect.TypeOf(WASMStateChangeItem{}), Name: "LogStateChangeItemWasm"},
		TyLogCreateUserWasmContract: {Ty: reflect.TypeOf(ReceiptWASMContract{}), Name: "LogCreateUserWasmContract"},
		TyLogOutputItemWasm: {Ty: reflect.TypeOf(WasmDebugResp{}), Name: "LogOutputItemWasm"},
	}
	return logInfo
}

