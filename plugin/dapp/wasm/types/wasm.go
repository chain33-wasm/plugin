// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"

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
	if err != nil {
		return tx.To
	}

	return ""
}

// Amount 获取金额
func (wasm WasmType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

// CreateTx 创建交易对象
func (wasm WasmType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	elog.Debug("wasm.CreateTx", "action", action)

	var param CreateOrCallWasmContract
	err := json.Unmarshal(message, &param)
	if err != nil {
		elog.Error("CreateTx", "Error", err)
		return nil, types.ErrInvalidParam
	}

	txType := param.Value.WasmContractActionType()
	if CreateWasmContractAction == txType || CallWasmContractAction == txType {
		return createWasmTx(&param)
	}

	return nil, types.ErrNotSupport
}

// GetLogMap 获取日志类型映射
func (wasm *WasmType) GetLogMap() map[int64]*types.LogInfo {
	logInfo := map[int64]*types.LogInfo{
		TyLogContractDataWasm:       {Ty: reflect.TypeOf(LogWASMContractData{}), Name: "LogContractDataWasm"},
		TyLogContractStateWasm:       {Ty: reflect.TypeOf(WASMContractState{}), Name: "LogContractStateWasm"},
		TyLogCallContractWasm:      {Ty: reflect.TypeOf(ReceiptWASMContract{}), Name: "LogCallContractWasm"},
		TyLogStateChangeItemWasm: {Ty: reflect.TypeOf(WASMStateChangeItem{}), Name: "LogStateChangeItemWasm"},
	}
	return logInfo
}

func createWasmTx(param *CreateOrCallWasmContract) (*types.Transaction, error) {
	txType := param.Value.WasmContractActionType()
	//创建部署wasm合约的交易
	if CreateWasmContractAction == txType {
		creatPara, ok := param.Value.(CreateWasmContractPara)
		if !ok {
			return nil, ErrCreateWasmPara
		}

		action := &WasmContractAction{
			Value: &WasmContractAction_CreateWasmContract{
				CreateWasmContract:&CreateWasmContract{
					GasLimit: uint64(creatPara.Fee),
					GasPrice:1,
					Code:  creatPara.Code,
					Abi:   creatPara.Abi,
					Name:  types.ExecName(creatPara.Name),
					Note:  creatPara.Note,
				},
			},
			Ty: CreateWasmContractAction,
		}

		return createRawWasmTx(action, WasmX, creatPara.Fee)
	}
    //创建调用用户自定义的user.wasm.xxx合约的交易
	callPara, ok := param.Value.(CallWasmContractPara)
	if !ok {
		return nil, ErrCreateWasmPara
	}



	action := &WasmContractAction{
		Value: &WasmContractAction_CallWasmContract{
			CallWasmContract:&CallWasmContract{
				GasLimit:uint64(callPara.Fee),
				GasPrice: 1,
				Note: callPara.Note,
				VmType: VMBinaryen, //当前只支持binaryen解释执行的方式
				ActionName:callPara.ActionName,
				ActionData:callPara.ActionData,
			},
		},
		Ty: CallWasmContractAction,
	}

	return createRawWasmTx(action, callPara.Name, callPara.Fee)
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
