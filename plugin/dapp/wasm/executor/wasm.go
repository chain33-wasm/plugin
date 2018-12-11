package executor

//#cgo CFLAGS: -I./../../../../wasmcpp/adapter/include
//#cgo LDFLAGS: -L./../../../../wasmcpp/lib -lwasm_adapter -ldl -lpthread -lz -ltinfo -lc  -lssl -lcrypto -lsecp256k1 -lchainbase -lWAST -lWASM -lRuntime -lIR -lLogging -lPlatform  -lwasm -lasmjs -lpasses -lcfg -last -lemscripten-optimizer -lsupport -lsoftfloat -lbuiltins -lfc -lm -lstdc++
//#cgo LDFLAGS: -L/usr/local/lib -lboost_filesystem -lboost_system -lboost_chrono -lboost_date_time
//#cgo LDFLAGS: -L/usr/lib/llvm-4.0/lib -lLLVMPasses -lLLVMipo -lLLVMInstrumentation -lLLVMVectorize -lLLVMIRReader -lLLVMAsmParser -lLLVMLinker -lLLVMMCJIT -lLLVMExecutionEngine -lLLVMRuntimeDyld -lLLVMDebugInfoDWARF -lLLVMX86CodeGen -lLLVMAsmPrinter -lLLVMDebugInfoCodeView -lLLVMDebugInfoMSF -lLLVMGlobalISel -lLLVMSelectionDAG -lLLVMCodeGen -lLLVMScalarOpts -lLLVMInstCombine -lLLVMBitWriter -lLLVMTransformUtils -lLLVMTarget -lLLVMAnalysis -lLLVMProfileData -lLLVMX86AsmParser -lLLVMX86Desc -lLLVMX86AsmPrinter -lLLVMX86Utils -lLLVMObject -lLLVMMCParser -lLLVMBitReader -lLLVMCore -lLLVMX86Disassembler -lLLVMX86Info -lLLVMMCDisassembler -lLLVMMC -lLLVMSupport -lLLVMDemangle
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <wasmcpp/adapter/include/eosio/chain/wasm_interface_adapter.h>
import "C"

import (
	"fmt"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
	"github.com/33cn/plugin/plugin/dapp/wasm/executor/state"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"unsafe"
	"bytes"
)

var (
	// 本合约地址
	wasmAddress    = address.ExecAddress(types.ExecName(wasmtypes.WasmX))
	pMemoryStateDB *state.MemoryStateDB
	pWasm          *WASMExecutor
	log            = log15.New("module", "execs.token")
)

func init() {
	ety := types.LoadExecutorType(wasmtypes.WasmX)
	ety.InitFuncList(types.ListMethod(&WASMExecutor{}))
}

func Init(name string, sub []byte) {
	drivers.Register(GetName(), newWASMDriver, 0)
	wasmAddress = address.ExecAddress(GetName())
}

func GetName() string {
	return types.ExecName(wasmtypes.WasmX)
}

func newWASMDriver() drivers.Driver {
	wasm := NewWASMExecutor()
	return wasm
}

// WASM执行器结构
type WASMExecutor struct {
	drivers.DriverBase
	vmCfg    *loccom.Config
	mStateDB *state.MemoryStateDB
	tx       *types.Transaction
	txIndex  int
	out2User []*wasmtypes.WasmOutItem
}

func NewWASMExecutor() *WASMExecutor {
	exec := &WASMExecutor{}

	exec.vmCfg = &loccom.Config{}
	exec.SetChild(exec)
	exec.SetExecutorType(types.LoadExecutorType(wasmtypes.WasmX))
	return exec
}

func setStateDB4Callback(memoryStateDB *state.MemoryStateDB) {
	pMemoryStateDB = memoryStateDB
}

func setWasm4Callback(wasm *WASMExecutor) {
	pWasm = wasm
	setStateDB4Callback(wasm.mStateDB)
}

//在获取key对应的value之前，需要先获取下value的size，为了避免传递的内存太小
//export StateDBGetValueSizeCallback
func StateDBGetValueSizeCallback(contractAddr *C.char, key *C.char, keyLen C.int) C.int {
	log.Debug("Entering StateDBGetValueSize")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	value := pMemoryStateDB.GetState(contractAddrgo, string(keySlice))
	return C.int(len(value))
}

//export StateDBGetStateCallback
func StateDBGetStateCallback(contractAddr *C.char, key *C.char, keyLen C.int, val *C.char, valLen C.int) C.int {
	log.Debug("Entering Get StateDBGetStateCallback")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	value := pMemoryStateDB.GetState(contractAddrgo, string(keySlice))
	if 0 == len(value) {
		log.Debug("Entering Get StateDBGetStateCallback", "get null value for key", string(keySlice))
		return 0
	}

	actualSize := len(value)
	if actualSize > int(valLen) {
		actualSize = int(valLen)
	}

	C.memcpy(unsafe.Pointer(val), unsafe.Pointer(&value[0]), C.size_t(actualSize))
	log.Debug("StateDBGetStateCallback", "key", string(keySlice), "value", value)

	return C.int(actualSize)
}

//export StateDBSetStateCallback
func StateDBSetStateCallback(contractAddr *C.char, key *C.char, keyLen C.int, val *C.char, valLen C.int) {
	log.Debug("Entering Set StateDBSetStateCallback")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	valueSlice := C.GoBytes(unsafe.Pointer(val), valLen)

	pMemoryStateDB.SetState(contractAddrgo, string(keySlice), valueSlice)

	log.Debug("StateDBSetStateCallback", "key", string(keySlice), "value in string:",
		"value in slice:", valueSlice)
}

//该接口用于返回查询结果的返回
//export Output2UserCallback
func Output2UserCallback(typeName *C.char, value *C.char, len C.int) {
	log.Debug("Entering Output2UserCallback")

	valueSlice := C.GoBytes(unsafe.Pointer(value), len)
	wasmOutItem := &wasmtypes.WasmOutItem{
		ItemType: C.GoString(typeName),
		Data:     valueSlice,
	}
	pWasm.out2User = append(pWasm.out2User, wasmOutItem)

	return
}

////////////以下接口用于user.wasm.xxx合约内部转账/////////////////////////////
//冻结user.wasm.xxx合约addr上的部分余额,其中的
//export ExecFrozen
func ExecFrozen(addr *C.char, amount C.longlong) int {
	return pWasm.mStateDB.ExecFrozen(pWasm.tx, C.GoString(addr), int64(amount))
}
//激活user.wasm.xxx合约addr上的部分余额
//export ExecActive
func ExecActive(addr *C.char, amount C.longlong) int {
	return pWasm.mStateDB.ExecActive(pWasm.tx, C.GoString(addr), int64(amount))
}

//export ExecTransfer
func ExecTransfer(from, to *C.char, amount C.longlong) int {
	return pWasm.mStateDB.ExecTransfer(pWasm.tx, C.GoString(from), C.GoString(to), int64(amount))
}

//export ExecTransferFrozen
func ExecTransferFrozen(from, to *C.char, amount C.longlong) int {
	return pWasm.mStateDB.ExecTransferFrozen(pWasm.tx, C.GoString(from), C.GoString(to), int64(amount))
}

func (wasm *WASMExecutor) GetName() string {
	return types.ExecName(wasmtypes.WasmX)
}

func (wasm *WASMExecutor) GetDriverName() string {
	return wasmtypes.WasmX
}

// Allow 允许哪些交易在本命执行器执行
func (wasm *WASMExecutor) Allow(tx *types.Transaction, index int) error {
	err := wasm.DriverBase.Allow(tx, index)
	if err == nil {
		return nil
	}
	//增加新的规则:
	//主链: user.wasm.xxx  执行 wasm用户自定义 合约
	//平行链: user.p.guodun.user.wasm.xxx 执行 wasm用户自定义合约
	exec := types.GetParaExec(tx.Execer)
	if wasm.AllowIsUserDot2(exec) {
		return nil
	}
	return types.ErrNotAllow
}

func (wasm *WASMExecutor) prepareExecContext(tx *types.Transaction, index int) {
	if wasm.mStateDB == nil {
		wasm.mStateDB = state.NewMemoryStateDB(types.ExecName(string(tx.Execer)), wasm.GetStateDB(), wasm.GetLocalDB(), wasm.GetCoinsAccount(), wasm.GetHeight())
	}

	wasm.tx = tx
	wasm.txIndex = index
}

func (wasm *WASMExecutor) prepareQueryContext(executorName string) {
	if wasm.mStateDB == nil {
		wasm.mStateDB = state.NewMemoryStateDB(executorName, wasm.GetStateDB(), wasm.GetLocalDB(), wasm.GetCoinsAccount(), wasm.GetHeight())
	}
}

// 根据交易hash生成一个新的合约对象地址
//func (wasm *WASMExecutor) getNewAddr(txHash []byte) *address.Address {
//	return address.GetExecAddress(loccom.CalcWasmContractName(txHash))
//}

func (wasm *WASMExecutor) GenerateExecReceipt(usedGas, gasPrice uint64, snapshot int, execName, caller, contractAddr string, opType loccom.WasmContratOpType) (*types.Receipt, error) {
	curVer := wasm.mStateDB.GetLastSnapshot()

	// 计算消耗了多少费用（实际消耗的费用）
	usedFee, overflow := loccom.SafeMul(usedGas, gasPrice)
	// 费用消耗溢出，执行失败
	if overflow || usedFee > uint64(wasm.tx.Fee) {
		// 如果操作没有回滚，则在这里处理
		if curVer != nil && snapshot >= curVer.GetId() && curVer.GetId() > -1 {
			wasm.mStateDB.RevertToSnapshot(snapshot)
		}
		log.Error("GenerateExecReceipt", "overflow", overflow, "usedfee", usedFee, "txFee", wasm.tx.Fee)
		return nil, wasmtypes.ErrOutOfGasWASM
	}

	// 打印合约中生成的日志
	wasm.mStateDB.PrintLogs()

	if curVer == nil {
		return nil, nil
	}
	// 从状态机中获取数据变更和变更日志
	data, logs := wasm.mStateDB.GetChangedData(curVer.GetId(), opType)
	contractReceipt := &wasmtypes.ReceiptWASMContract{caller, execName, contractAddr, usedGas}

	runLog := &types.ReceiptLog{
		Ty:  wasmtypes.TyLogCallContractWasm,
		Log: types.Encode(contractReceipt)}
	if opType == wasmtypes.CreateWasmContractAction {
		runLog.Ty = wasmtypes.TyLogCreateUserWasmContract
	}

	//调用wasm执行器的log
	logs = append(logs, runLog)
	logs = append(logs, wasm.mStateDB.GetReceiptLogs(contractAddr)...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: data, Logs: logs}

	// 返回之前，把本次交易在区块中生成的合约日志集中打印出来
	if wasm.mStateDB != nil {
		wasm.mStateDB.WritePreimages(wasm.GetHeight())
	}

	wasm.collectWasmTxLog(wasm.tx, contractReceipt, receipt)

	return receipt, nil
}

func (wasm *WASMExecutor) queryFromExec(contractAddr string,
	actionName string,
	actionData []byte) ([]*wasmtypes.WasmOutItem, error) {

	code := wasm.mStateDB.GetCode(contractAddr)
	if nil == code {
		log.Error("call wasm contract ", "failed to get code from contract address", contractAddr)
		return nil, wasmtypes.ErrWrongContractAddr
	}
	AliasStr := wasm.mStateDB.GetName(contractAddr)

	setWasm4Callback(wasm)

	actiondata4C := C.CBytes(actionData)
	ContractAddr := C.CString(contractAddr)
	Alias := C.CString(AliasStr)
	ActionName := C.CString(actionName)
	from := C.CString(address.PubKeyToAddress(wasm.tx.GetSignature().GetPubkey()).String())
	defer C.free(unsafe.Pointer(actiondata4C))
	defer C.free(unsafe.Pointer(Alias))
	defer C.free(unsafe.Pointer(ContractAddr))
	defer C.free(unsafe.Pointer(ActionName))
	defer C.free(unsafe.Pointer(from))

	context := &C.Apply_context_para{
		contractAddr: ContractAddr,
		contractName: Alias,
		action_name:  ActionName,
		pdata:        (*C.char)(actiondata4C),
		datalen:      C.int(len(actionData)),
		from:         from,
		gasAvailable: C.int64_t(0x100000),
		blocktime:    C.int64_t(wasm.GetBlockTime()),
		height:       C.int64_t(wasm.GetHeight()),
	}

	//2nd step: just call contract
	codePtr := C.CBytes(code)
	C.callContract4go(C.VMTypeBinaryen, (*C.char)(codePtr), C.int(len(code)), context)
	defer C.free(codePtr)

	return wasm.out2User, nil
}

func (wasm *WASMExecutor) collectWasmTxLog(tx *types.Transaction, cr *wasmtypes.ReceiptWASMContract, receipt *types.Receipt) {
	log.Debug("wasm collect begin")
	log.Debug("Tx info", "txHash", common.Bytes2Hex(tx.Hash()), "height", wasm.GetHeight())
	log.Debug("ReceiptWASMContract", "data", fmt.Sprintf("caller=%v, name=%v, addr=%v, usedGas=%v", cr.Caller, cr.ContractName, cr.ContractAddr, cr.UsedGas))
	log.Debug("receipt data", "type", receipt.Ty)
	for _, kv := range receipt.KV {
		log.Debug("KeyValue", "key", common.Bytes2Hex(kv.Key), "value", common.Bytes2Hex(kv.Value))
	}
	for _, kv := range receipt.Logs {
		log.Debug("ReceiptLog", "Type", kv.Ty, "log", common.Bytes2Hex(kv.Log))
	}
	log.Debug("wasm collect end")
}

func (wasm *WASMExecutor) ExecLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := wasm.DriverBase.ExecLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}

	// 需要将Exec中生成的合约状态变更信息写入localdb
	for _, logItem := range receipt.Logs {
		if wasmtypes.TyLogStateChangeItemWasm == logItem.Ty {
			data := logItem.Log
			var changeItem wasmtypes.WASMStateChangeItem
			err = types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}
			set.KV = append(set.KV, &types.KeyValue{Key: []byte(changeItem.Key), Value: changeItem.CurrentValue})
		}
	}

	return set, err
}

func (wasm *WASMExecutor) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := wasm.DriverBase.ExecDelLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}

	// 需要将Exec中生成的合约状态变更信息从localdb中恢复
	for _, logItem := range receipt.Logs {
		if wasmtypes.TyLogStateChangeItemWasm == logItem.Ty {
			data := logItem.Log
			var changeItem wasmtypes.WASMStateChangeItem
			err = types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}
			set.KV = append(set.KV, &types.KeyValue{Key: []byte(changeItem.Key), Value: changeItem.PreValue})
		}
	}

	return set, err
}

func (wasm *WASMExecutor) getContractTable(in *wasmtypes.WasmQueryContractTableReq) (types.Message, error) {
	resp := &wasmtypes.WasmQueryResponse{}
	contractAddr := address.ExecAddress(types.ExecName(in.ContractName))
	wasm.prepareQueryContext(types.ExecName(wasmtypes.WasmX))
	abi := wasm.mStateDB.GetAbi(contractAddr)
	if nil == abi {
		log.Error("getContractTable", "Failed to get abi for wasm contract", in.ContractName)
		return nil, wasmtypes.ErrAddrNotExists
	}
	wasm.mStateDB.SetCurrentExecutorName(types.ExecName(in.ContractName))
	abi4CStr := C.CString(string(abi))
	defer C.free(unsafe.Pointer(abi4CStr))

	var wasmOutItems []*wasmtypes.WasmOutItem
	for _, item := range in.Items {
		data := wasm.mStateDB.GetState(contractAddr, item.Key)
		wasmOutItem := &wasmtypes.WasmOutItem{
			ItemType: item.TableName,
			Data:data,
		}
		wasmOutItems = append(wasmOutItems, wasmOutItem)
	}

	for _, wasmOutItem := range wasmOutItems {
		if nil == wasmOutItem.Data {
			result := &wasmtypes.QueryResultItem{
				ItemType:   wasmOutItem.ItemType,
				ResultJSON: "Error:can't find this kind of table",
				Found:      false,
			}
			resp.QueryResultItems = append(resp.QueryResultItems, result)
			continue
		}
		var jsonResult *C.char
		structName := C.CString(wasmOutItem.ItemType)
		serializedData := C.CBytes(wasmOutItem.Data)

		log.Debug("wasm query", "structure", wasmOutItem.ItemType)

		if 0 != C.convertData2Json(abi4CStr, (*C.char)(serializedData), (C.int)(len(wasmOutItem.Data)), structName, &jsonResult) {
			log.Error("wasm query", "structure", wasmOutItem.ItemType)
			return nil, wasmtypes.ErrUnserialize
		}

		result := &wasmtypes.QueryResultItem{
			ItemType:   wasmOutItem.ItemType,
			ResultJSON: C.GoString(jsonResult),
			Found:      true,
		}

		log.Debug("wasm query", "ResultJSON", result.ResultJSON)

		resp.QueryResultItems = append(resp.QueryResultItems, result)
		C.free(unsafe.Pointer(structName))
		C.free(unsafe.Pointer(serializedData))
		C.free(unsafe.Pointer(jsonResult))
	}

	return resp, nil
}

// 检查合约地址是否存在，此操作不会改变任何状态，所以可以直接从statedb查询
func (wasm *WASMExecutor) checkContractNameExists(req *wasmtypes.CheckWASMContractNameReq) (types.Message, error) {
	contractName := req.WasmContractName
	if len(contractName) == 0 {
		return nil, wasmtypes.ErrAddrNotExists
	}

	if !bytes.Contains([]byte(contractName), []byte(wasmtypes.UserWasmX)) {
		contractName = wasmtypes.UserWasmX + contractName
	}

	exists := wasm.GetMStateDB().Exist(address.ExecAddress(contractName))
	ret := &wasmtypes.CheckWASMAddrResp{ExistAlready: exists}
	return ret, nil
}


func (wasm *WASMExecutor) GetMStateDB() *state.MemoryStateDB {
	return wasm.mStateDB
}

// 从交易信息中获取交易目标地址，在创建合约交易中，此地址为空
func getReceiver(tx *types.Transaction) *address.Address {
	if tx.To == "" {
		return nil
	}

	addr, err := address.NewAddrFromString(tx.To)
	if err != nil {
		log.Error("create address form string error", "string:", tx.To)
		return nil
	}

	return addr
}

// 检查合约调用账户是否有充足的金额进行转账交易操作
func CanTransfer(db state.StateDB, sender, recipient address.Address, amount uint64) bool {
	return db.CanTransfer(sender.String(), recipient.String(), amount)
}

// 在内存数据库中执行转账操作（只修改内存中的金额）
// 从外部账户地址到合约账户地址
func Transfer(db state.StateDB, sender, recipient address.Address, amount uint64) bool {
	return db.Transfer(sender.String(), recipient.String(), amount)
}