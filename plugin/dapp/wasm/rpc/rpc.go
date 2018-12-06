// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

// CreateWasmContractTx 创建未签名的创建wasm合约的交易
func (c *Jrpc) CreateWasmContractTx(param *wasmtypes.CreateWasmContract, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(wasmtypes.WasmX), "CreateWasmContract", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// CallWasmContractTx 创建未签名的调用wasm合约的交易
func (c *Jrpc) CallWasmContractTx(param *wasmtypes.CallWasmContract, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(wasmtypes.WasmX), "CallWasmContract", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}
