package executor

import (
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

// ExecLocal_CreateWasmContract : 本地执行创建wasm合约
func (wasm *WASMExecutor) ExecLocal_CreateWasmContract(payload *wasmtypes.CreateWasmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return wasm.execLocal(tx, receipt, index)
}

// ExecLocal_CallWasmContract : 本地执行调用wasm合约
func (wasm *WASMExecutor) ExecLocal_CallWasmContract(payload *wasmtypes.CallWasmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return wasm.execLocal(tx, receipt, index)
}

func (wasm *WASMExecutor) execLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	// 需要将Exec中生成的合约状态变更信息写入localdb
	for _, logItem := range receipt.Logs {
		if wasmtypes.TyLogStateChangeItemWasm == logItem.Ty {
			data := logItem.Log
			var changeItem wasmtypes.WASMStateChangeItem
			err := types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}
			set.KV = append(set.KV, &types.KeyValue{Key: []byte(changeItem.Key), Value: changeItem.CurrentValue})
		}
	}

	return set, nil
}
