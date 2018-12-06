package executor

//#cgo CFLAGS: -I./../../../../wasmcpp/adapter/include
//#cgo LDFLAGS: -L./../../../../wasmcpp/lib -lwasm_adapter -ldl -lpthread -lz -ltinfo -lc  -lssl -lcrypto -lsecp256k1 -lchainbase -lWAST -lWASM -lRuntime -lIR -lLogging -lPlatform  -lwasm -lasmjs -lpasses -lcfg -last -lemscripten-optimizer -lsupport -lsoftfloat -lbuiltins -lfc -lm -lstdc++
//#cgo LDFLAGS: -L/usr/local/lib -lboost_filesystem -lboost_system -lboost_chrono -lboost_date_time
//#cgo LDFLAGS: -L/usr/lib/llvm-4.0/lib -lLLVMPasses -lLLVMipo -lLLVMInstrumentation -lLLVMVectorize -lLLVMIRReader -lLLVMAsmParser -lLLVMLinker -lLLVMMCJIT -lLLVMExecutionEngine -lLLVMRuntimeDyld -lLLVMDebugInfoDWARF -lLLVMX86CodeGen -lLLVMAsmPrinter -lLLVMDebugInfoCodeView -lLLVMDebugInfoMSF -lLLVMGlobalISel -lLLVMSelectionDAG -lLLVMCodeGen -lLLVMScalarOpts -lLLVMInstCombine -lLLVMBitWriter -lLLVMTransformUtils -lLLVMTarget -lLLVMAnalysis -lLLVMProfileData -lLLVMX86AsmParser -lLLVMX86Desc -lLLVMX86AsmPrinter -lLLVMX86Utils -lLLVMObject -lLLVMMCParser -lLLVMBitReader -lLLVMCore -lLLVMX86Disassembler -lLLVMX86Info -lLLVMMCDisassembler -lLLVMMC -lLLVMSupport -lLLVMDemangle
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <../../../../wasmcpp/adapter/include/eosio/chain/wasm_interface_adapter.h>
import "C"

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"unsafe"
)

func (wasm *WASMExecutor) Exec_CreateWasmContract(CreateWasmContract *wasmtypes.CreateWasmContract) (*types.Receipt, error) {
	// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）
	contractAddr := wasm.getNewAddr(wasm.tx.Hash())
	contractAddrInStr := contractAddr.String()
	if !wasm.mStateDB.Empty(contractAddrInStr) {
		return nil, loccom.ErrContractAddressCollisionWASM
	}

	log.Debug("wasm create", "new created wasm contract addr =", contractAddrInStr)

	codeSize := len(CreateWasmContract.GetCode())
	if codeSize > loccom.MaxCodeSize {
		return nil, loccom.ErrMaxCodeSizeExceededWASM
	}

	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := address.PubKeyToAddress(wasm.tx.GetSignature().GetPubkey())
	to := getReceiver(wasm.tx)
	if to == nil {
		return nil, types.ErrInvalidAddress
	}

	snapshot := wasm.mStateDB.Snapshot()
	//验证合约代码的正确性
	code := C.CBytes(CreateWasmContract.Code)
	defer C.free(code)
	if result := C.wasm_validate_contract((*C.char)(code), C.int(len(CreateWasmContract.Code))); result != C.Success {
		log.Error("wasm_validate_contract", "failed with result", result)
		return nil, loccom.ErrWASMValidationFail
	}
	// 创建新的合约对象，包含双方地址以及合约代码，可用Gas信息
	contract := loccom.NewContract(loccom.AccountRef(*from), loccom.AccountRef(*contractAddr), 0, CreateWasmContract.GasLimit)
	contract.SetCallCode(*contractAddr, common.BytesToHash(common.Sha256(CreateWasmContract.Code)), CreateWasmContract.Code)

	// 创建一个新的账户对象（合约账户）
	execName := loccom.CalcWasmContractName(wasm.tx.Hash())
	wasm.mStateDB.CreateAccount(contractAddrInStr, contract.CallerAddress.String(), execName, CreateWasmContract.Alias)

	createDataGas := (uint64(len(CreateWasmContract.Code)) + uint64(len(CreateWasmContract.Abi))) * loccom.CreateDataGas
	if contract.UseGas(createDataGas) {
		wasm.mStateDB.SetCodeAndAbi(contractAddrInStr, CreateWasmContract.Code, []byte(CreateWasmContract.Abi))
	} else {
		return nil, loccom.ErrCodeStoreOutOfGasWASM
	}

	usedGas := CreateWasmContract.GasLimit - contract.Gas

	receipt, err := wasm.GenerateExecReceipt(usedGas,
		uint64(CreateWasmContract.GasPrice),
		snapshot,
		execName,
		contract.CallerAddress.String(),
		contractAddrInStr,
		loccom.CreateWasmContrcat)
	log.Debug("wasm create", "receipt", receipt, "err info", err)

	return receipt, err
}

func (wasm *WASMExecutor) Exec_CallWasmContract(callWasmContract *wasmtypes.CallWasmContract) (*types.Receipt, error) {
	if callWasmContract.VmType != wasmtypes.VMBinaryen {
		panic("Now only binaryen is supported")
		return nil, loccom.ErrWASMWavmNotSupported
	}

	log.Debug("wasm call", "Para CallWasmContract", callWasmContract)

	code := wasm.mStateDB.GetCode(callWasmContract.ContractAddr)
	if nil == code {
		log.Error("call wasm contract ", "failed to get code from contract address", callWasmContract.ContractAddr)
		return nil, loccom.ErrWrongContractAddr
	}

	snapshot := wasm.mStateDB.Snapshot()
	setWasm4Callback(wasm)

	//1st step: create apply context
	log.Debug("wasm call para", "ActionData", callWasmContract.ActionData,
		"ContractAddr", callWasmContract.ContractAddr,
		"Alias", callWasmContract.Alias,
		"ActionName", callWasmContract.ActionName)
	actiondata := C.CBytes(callWasmContract.ActionData)
	ContractAddr := C.CString(callWasmContract.ContractAddr)
	Alias := C.CString(callWasmContract.Alias)
	ActionName := C.CString(callWasmContract.ActionName)
	from := C.CString(address.PubKeyToAddress(wasm.tx.GetSignature().GetPubkey()).String())
	defer C.free(unsafe.Pointer(actiondata))
	defer C.free(unsafe.Pointer(ContractAddr))
	defer C.free(unsafe.Pointer(Alias))
	defer C.free(unsafe.Pointer(ActionName))
	defer C.free(unsafe.Pointer(from))

	context := &C.Apply_context_para{
		contractAddr: ContractAddr,
		contractName: Alias,
		action_name:  ActionName,
		pdata:        (*C.char)(actiondata),
		datalen:      C.int(len(callWasmContract.ActionData)),
		from:         from,
		gasAvailable: C.int64_t(callWasmContract.GasLimit),
		blocktime:    C.int64_t(wasm.GetBlockTime()),
		height:       C.int64_t(wasm.GetHeight()),
	}

	//2nd step: just call contract
	codePtr := C.CBytes(code)
	leftGas := C.callContract4go(C.VMTypeBinaryen, (*C.char)(codePtr), C.int(len(code)), context)
	defer C.free(codePtr)
	log.Debug("wasm call", "call back from callContract4go with leftGas", leftGas)

	//合约执行失败
	if leftGas < 0 && leftGas == wasmtypes.GAS_EXHAUSTED_ERR_CODE {
		wasm.mStateDB.RevertToSnapshot(snapshot)
		log.Error("call wasm contract ", "failed to call contract due to", loccom.ErrWasmContractExecFailed)
		return nil, loccom.ErrWasmContractExecFailed
	} else if leftGas < 0 {
		//合约购买的gas不够
		wasm.mStateDB.RevertToSnapshot(snapshot)
		log.Error("call wasm contract ", "failed to call contract due to", loccom.ErrOutOfGasWASM)
		return nil, loccom.ErrOutOfGasWASM
	}
	usedGas := callWasmContract.GasLimit - uint64(leftGas)

	contractAccount := wasm.mStateDB.GetAccount(callWasmContract.ContractAddr)
	caller := address.PubKeyToAddress(wasm.tx.GetSignature().GetPubkey()).String()

	receipt, err := wasm.GenerateExecReceipt(usedGas,
		uint64(callWasmContract.GasPrice),
		snapshot,
		contractAccount.GetExecName(),
		caller,
		callWasmContract.ContractAddr,
		loccom.CallWasmContrcat)
	log.Debug("wasm call", "receipt", receipt, "err info", err)

	return receipt, err
}
