package executor

//#cgo CFLAGS: -I./wasmcpp/adapter/include
//#cgo LDFLAGS: -L./wasmcpp/lib -lwasm_adapter -ldl -lpthread -lz -ltinfo -lc  -lssl -lcrypto -lsecp256k1 -lchainbase -lWAST -lWASM -lRuntime -lIR -lLogging -lPlatform  -lwasm -lasmjs -lpasses -lcfg -last -lemscripten-optimizer -lsupport -lsoftfloat -lbuiltins -lfc -lm -lstdc++
//#cgo LDFLAGS: -L/usr/local/lib -lboost_filesystem -lboost_system -lboost_chrono -lboost_date_time
//#cgo LDFLAGS: -L/usr/lib/llvm-4.0/lib -lLLVMPasses -lLLVMipo -lLLVMInstrumentation -lLLVMVectorize -lLLVMIRReader -lLLVMAsmParser -lLLVMLinker -lLLVMMCJIT -lLLVMExecutionEngine -lLLVMRuntimeDyld -lLLVMDebugInfoDWARF -lLLVMX86CodeGen -lLLVMAsmPrinter -lLLVMDebugInfoCodeView -lLLVMDebugInfoMSF -lLLVMGlobalISel -lLLVMSelectionDAG -lLLVMCodeGen -lLLVMScalarOpts -lLLVMInstCombine -lLLVMBitWriter -lLLVMTransformUtils -lLLVMTarget -lLLVMAnalysis -lLLVMProfileData -lLLVMX86AsmParser -lLLVMX86Desc -lLLVMX86AsmPrinter -lLLVMX86Utils -lLLVMObject -lLLVMMCParser -lLLVMBitReader -lLLVMCore -lLLVMX86Disassembler -lLLVMX86Info -lLLVMMCDisassembler -lLLVMMC -lLLVMSupport -lLLVMDemangle
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <wasmcpp/adapter/include/eosio/chain/wasm_interface_adapter.h>
import "C"

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"unsafe"
)

func (wasm *WASMExecutor) Exec_CreateWasmContract(createWasmContract *wasmtypes.CreateWasmContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	wasm.prepareExecContext(tx, index)
	// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）
	contractAddr := address.GetExecAddress(createWasmContract.Name)
	contractAddrInStr := contractAddr.String()
	if !wasm.mStateDB.Empty(contractAddrInStr) {
		return nil, wasmtypes.ErrContractAddressCollisionWASM
	}

	log.Debug("wasm create", "new created wasm contract addr =", contractAddrInStr)

	codeSize := len(createWasmContract.GetCode())
	if codeSize > loccom.MaxCodeSize {
		return nil, wasmtypes.ErrMaxCodeSizeExceededWASM
	}

	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := address.PubKeyToAddress(wasm.tx.GetSignature().GetPubkey())
	to := getReceiver(wasm.tx)
	if to == nil {
		return nil, types.ErrInvalidAddress
	}

	snapshot := wasm.mStateDB.Snapshot()
	//验证合约代码的正确性
	code := C.CBytes(createWasmContract.Code)
	defer C.free(code)
	if result := C.wasm_validate_contract((*C.char)(code), C.int(len(createWasmContract.Code))); result != C.Success {
		log.Error("wasm_validate_contract", "failed with result", result)
		return nil, wasmtypes.ErrWASMValidationFail
	}
	// 创建新的合约对象，包含双方地址以及合约代码，可用Gas信息
	contract := loccom.NewContract(loccom.AccountRef(*from), loccom.AccountRef(*contractAddr), 0, createWasmContract.GasLimit)
	contract.SetCallCode(*contractAddr, common.BytesToHash(common.Sha256(createWasmContract.Code)), createWasmContract.Code)

	// 创建一个新的账户对象（合约账户）
	wasm.mStateDB.CreateAccount(contractAddrInStr, contract.CallerAddress.String(), createWasmContract.Name)

	createDataGas := (uint64(len(createWasmContract.Code)) + uint64(len(createWasmContract.Abi))) * loccom.CreateDataGas
	if contract.UseGas(createDataGas) {
		wasm.mStateDB.SetCodeAndAbi(contractAddrInStr, createWasmContract.Code, []byte(createWasmContract.Abi))
	} else {
		return nil, wasmtypes.ErrCodeStoreOutOfGasWASM
	}

	usedGas := createWasmContract.GasLimit - contract.Gas

	receipt, err := wasm.GenerateExecReceipt(usedGas,
		uint64(createWasmContract.GasPrice),
		snapshot,
		createWasmContract.Name,
		contract.CallerAddress.String(),
		contractAddrInStr,
		wasmtypes.CreateWasmContractAction)
	log.Debug("wasm create", "receipt", receipt, "err info", err)

	return receipt, err
}

func (wasm *WASMExecutor) Exec_CallWasmContract(callWasmContract *wasmtypes.CallWasmContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	wasm.prepareExecContext(tx, index)
	//因为在真正地执行user.wasm.xxx合约前，还需要通过wasm合约平台获取其合约字节码，
	//所以需要先将其合约名字设置为wasm
	wasm.mStateDB.SetCurrentExecutorName(wasmtypes.WasmX)
	if callWasmContract.VmType != wasmtypes.VMBinaryen {
		panic("Now only binaryen is supported")
		return nil, wasmtypes.ErrWASMWavmNotSupported
	}

	log.Debug("wasm call", "Para CallWasmContract", callWasmContract)

	userWasmAddr := address.ExecAddress(string(tx.Execer))
	code := wasm.mStateDB.GetCode(userWasmAddr)
	if nil == code {
		log.Error("call wasm contract ", "failed to get code from contract", string(tx.Execer))
		return nil, wasmtypes.ErrWrongContractAddr
	}
	//将当前合约执行名字修改为user.wasm.xxx
	wasm.mStateDB.SetCurrentExecutorName(string(types.GetParaExec(tx.Execer)))
	snapshot := wasm.mStateDB.Snapshot()
	setWasm4Callback(wasm)

	//1st step: create apply context
	log.Debug("wasm call para", "ActionData", callWasmContract.ActionData,
		"ContractName", string(tx.Execer),
		"ActionName", callWasmContract.ActionName)
	actiondata := C.CBytes(callWasmContract.ActionData)
	ContractAddr := C.CString(userWasmAddr)
	Alias := C.CString(string(tx.Execer))
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
		log.Error("call wasm contract ", "failed to call contract due to", wasmtypes.ErrWasmContractExecFailed)
		return nil, wasmtypes.ErrWasmContractExecFailed
	} else if leftGas < 0 {
		//合约购买的gas不够
		wasm.mStateDB.RevertToSnapshot(snapshot)
		log.Error("call wasm contract ", "failed to call contract due to", wasmtypes.ErrOutOfGasWASM)
		return nil, wasmtypes.ErrOutOfGasWASM
	}
	usedGas := callWasmContract.GasLimit - uint64(leftGas)

	contractAccount := wasm.mStateDB.GetAccount(userWasmAddr)
	caller := tx.From()


	receipt, err := wasm.GenerateExecReceipt(usedGas,
		uint64(callWasmContract.GasPrice),
		snapshot,
		contractAccount.GetExecName(),
		caller,
		userWasmAddr,
		wasmtypes.CallWasmContractAction)
	log.Debug("wasm call", "receipt", receipt, "err info", err)

	return receipt, err
}
