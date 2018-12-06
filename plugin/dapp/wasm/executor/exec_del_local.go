package executor

import (
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

// ExecDelLocal_CreateWasmContract : 本地撤销执行创建wasm合约
func (wasm *WASMExecutor) ExecDelLocal_CreateWasmContract(payload *wasmtypes.CreateWasmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return wasm.execDelLocal(tx, receipt, index)
}

// ExecDelLocal_CallWasmContract : 本地撤销执行调用wasm合约
func (wasm *WASMExecutor) ExecDelLocal_CallWasmContract(payload *wasmtypes.CallWasmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return wasm.execDelLocal(tx, receipt, index)
}

func (wasm *WASMExecutor) execDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	// 需要将Exec中生成的合约状态变更信息从localdb中恢复
	for _, logItem := range receipt.Logs {
		if wasmtypes.TyLogStateChangeItemWasm == logItem.Ty {
			data := logItem.Log
			var changeItem wasmtypes.WASMStateChangeItem
			err := types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}
			set.KV = append(set.KV, &types.KeyValue{Key: []byte(changeItem.Key), Value: changeItem.PreValue})
		}
	}

	return set, nil
}
