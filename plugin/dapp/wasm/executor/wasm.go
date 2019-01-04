package executor

//#cgo CFLAGS: -I./../../../../wasmcpp/adapter/include
//#cgo LDFLAGS: -L./../../../../wasmcpp/lib -lwasm_adapter -ldl -lpthread -lz -ltinfo -lc -lssl -lcrypto -lchainbase -lWAST -lWASM -lRuntime -lIR -lLogging -lPlatform  -lwasm -lasmjs -lpasses -lcfg -last -lemscripten-optimizer -lsupport -lsoftfloat -lbuiltins -lfc -lm -lstdc++
//#cgo LDFLAGS: -L/usr/local/lib -lboost_filesystem -lboost_system -lboost_chrono -lboost_date_time -lsecp256k1 -lgmp
//#cgo LDFLAGS: -L/usr/lib/llvm-4.0/lib -lLLVMPasses -lLLVMipo -lLLVMInstrumentation -lLLVMVectorize -lLLVMIRReader -lLLVMAsmParser -lLLVMLinker -lLLVMMCJIT -lLLVMExecutionEngine -lLLVMRuntimeDyld -lLLVMDebugInfoDWARF -lLLVMX86CodeGen -lLLVMAsmPrinter -lLLVMDebugInfoCodeView -lLLVMDebugInfoMSF -lLLVMGlobalISel -lLLVMSelectionDAG -lLLVMCodeGen -lLLVMScalarOpts -lLLVMInstCombine -lLLVMBitWriter -lLLVMTransformUtils -lLLVMTarget -lLLVMAnalysis -lLLVMProfileData -lLLVMX86AsmParser -lLLVMX86Desc -lLLVMX86AsmPrinter -lLLVMX86Utils -lLLVMObject -lLLVMMCParser -lLLVMBitReader -lLLVMCore -lLLVMX86Disassembler -lLLVMX86Info -lLLVMMCDisassembler -lLLVMMC -lLLVMSupport -lLLVMDemangle
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <wasmcpp/adapter/include/eosio/chain/wasm_interface_adapter.h>
import "C"

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
	"github.com/33cn/plugin/plugin/dapp/wasm/executor/state"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"google.golang.org/grpc"
	"context"
	"unsafe"
	"time"
)

type subConfig struct {
	ParaRemoteGrpcClient string `json:"paraRemoteGrpcClient"`
}

var (
	// 本合约地址
	wasmAddress    = address.ExecAddress(types.ExecName(wasmtypes.WasmX))
	pMemoryStateDB *state.MemoryStateDB
	pWasm          *WASMExecutor
	log            = log15.New("module", "execs.wasm")
    cfg subConfig
)

func init() {
	ety := types.LoadExecutorType(wasmtypes.WasmX)
	ety.InitFuncList(types.ListMethod(&WASMExecutor{}))
}

func Init(name string, sub []byte) {
	drivers.Register(wasmtypes.WasmX, newWASMDriver, 0)
	wasmAddress = address.ExecAddress(wasmtypes.WasmX)

	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
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
	grpcClient   types.Chain33Client
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

/////////////////////////LocalDB interface//////////////////////////////////////////
//export GetValueSizeFromLocal
func GetValueSizeFromLocal(contractAddr *C.char, key *C.char, keyLen C.int) C.int {
	log.Debug("Entering GetValueSizeFromLocal")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	value := pMemoryStateDB.GetValueFromLocal(contractAddrgo, string(keySlice))
	return C.int(len(value))
}

//export GetValueFromLocal
func GetValueFromLocal(contractAddr *C.char, key *C.char, keyLen C.int, val *C.char, valLen C.int) C.int {
	log.Debug("Entering GetValueFromLocal")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	value := pMemoryStateDB.GetValueFromLocal(contractAddrgo, string(keySlice))
	if 0 == len(value) {
		log.Debug("Entering Get StateDBGetStateCallback", "get null value for key", string(keySlice))
		return 0
	}

	actualSize := len(value)
	//不超出需要获取的数据的空间，以免导致内存越界
	if actualSize > int(valLen) {
		actualSize = int(valLen)
	}

	C.memcpy(unsafe.Pointer(val), unsafe.Pointer(&value[0]), C.size_t(actualSize))
	log.Debug("StateDBGetStateCallback", "key", string(keySlice), "value", value)

	return C.int(actualSize)
}

//export SetValue2Local
func SetValue2Local(contractAddr *C.char, key *C.char, keyLen C.int, val *C.char, valLen C.int) {
	log.Debug("Entering SetValue2Local")
	contractAddrgo := C.GoString(contractAddr)
	keySlice := C.GoBytes(unsafe.Pointer(key), keyLen)
	valueSlice := C.GoBytes(unsafe.Pointer(val), valLen)

	pMemoryStateDB.SetValue2Local(contractAddrgo, string(keySlice), valueSlice)

	log.Debug("StateDBSetStateCallback", "key", string(keySlice), "value in string:",
		"value in slice:", valueSlice)
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
	//不超出需要获取的数据的空间，以免导致内存越界
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
func ExecFrozen(addr *C.char, amount C.longlong) C.int {
	if nil == pWasm || nil == pWasm.mStateDB {
		log.Error("ExecFrozen failed due to nil handle", "pWasm", pWasm, "pWasm.mStateDB", pWasm.mStateDB)
		return C.int(wasmtypes.AccountOpFail)
	}
	return C.int(pWasm.mStateDB.ExecFrozen(pWasm.tx, C.GoString(addr), int64(int64(amount) * wasmtypes.Coin_Precision)))
}

//激活user.wasm.xxx合约addr上的部分余额
//export ExecActive
func ExecActive(addr *C.char, amount C.longlong) C.int {
	if nil == pWasm || nil == pWasm.mStateDB {
		log.Error("ExecActive failed due to nil handle", "pWasm", pWasm, "pWasm.mStateDB", pWasm.mStateDB)
		return C.int(wasmtypes.AccountOpFail)
	}
	return C.int(pWasm.mStateDB.ExecActive(pWasm.tx, C.GoString(addr), int64(int64(amount) * wasmtypes.Coin_Precision)))
}

//export ExecTransfer
func ExecTransfer(from, to *C.char, amount C.longlong) C.int {
	if nil == pWasm || nil == pWasm.mStateDB {
		log.Error("ExecTransfer failed due to nil handle", "pWasm", pWasm, "pWasm.mStateDB", pWasm.mStateDB)
		return C.int(wasmtypes.AccountOpFail)
	}
	return C.int(pWasm.mStateDB.ExecTransfer(pWasm.tx, C.GoString(from), C.GoString(to), int64(int64(amount) * wasmtypes.Coin_Precision)))
}

//export ExecTransferFrozen
func ExecTransferFrozen(from, to *C.char, amount C.longlong) C.int {
	if nil == pWasm || nil == pWasm.mStateDB {
		log.Error("ExecTransferFrozen failed due to nil handle", "pWasm", pWasm, "pWasm.mStateDB", pWasm.mStateDB)
		return C.int(wasmtypes.AccountOpFail)
	}
	return C.int(pWasm.mStateDB.ExecTransferFrozen(pWasm.tx, C.GoString(from), C.GoString(to), int64(int64(amount) * wasmtypes.Coin_Precision)))
}

//为wasm用户自定义合约提供随机数，该随机数是64位hash值,返回值为实际返回的长度
//export GetRandom
func GetRandom(randomDataOutput *C.char, maxLen C.int) C.int {
	var msg types.Message
	var err error
	var hash []byte
	blockNum := int64(5)
	if nil == pWasm {
		log.Error("GetRandom failed due to nil handle", "pWasm", pWasm)
		return C.int(wasmtypes.AccountOpFail)
	}

	//发消息给random模块
	//在主链上，当前高度查询不到，如果要保证区块个数，高度传入action.height-1
	log.Debug("GetRandom")
	if !types.IsPara() {
		req := &types.ReqRandHash{ExecName: "ticket", Height: pWasm.GetHeight() - 1, BlockNum: blockNum}
		msg, err = pWasm.GetAPI().Query("ticket", "RandNumHash", req)
		if err != nil {
			return -1
		}
		reply := msg.(*types.ReplyHash)
		hash = reply.Hash
	} else {
		mainHeight := pWasm.GetMainHeightByTxHash(pWasm.tx.Hash())
		if mainHeight < 0 {
			log.Error("GetRandom", "mainHeight", mainHeight)
			return -1
		}
		req := &types.ReqRandHash{ExecName: "ticket", Height: mainHeight, BlockNum: blockNum}
		reply, err := pWasm.grpcClient.QueryRandNum(context.Background(), req)
		if err != nil {
			return -1
		}
		hash = reply.Hash
	}
	randLen := C.int(len(hash))
	if randLen > maxLen {
		randLen = maxLen
	}
	//random := C.GoBytes(unsafe.Pointer(randomDataOutput), maxLen)
	C.memcpy(unsafe.Pointer(randomDataOutput), unsafe.Pointer(&hash[0]), C.size_t(randLen))
	//temp := C.int(copy(random, hash))

	return C.int(randLen)
}

func (wasm *WASMExecutor) GetMainHeightByTxHash(txHash []byte) int64 {
	for i := 0; i < wasmtypes.RetryNum; i++ {
		req := &types.ReqHash{Hash: txHash}
		txDetail, err := pWasm.grpcClient.QueryTransaction(context.Background(), req)
		if err != nil {
			time.Sleep(time.Second)
		} else {
			return txDetail.GetHeight()
		}
	}

	return -1
}

func (wasm *WASMExecutor) GetName() string {
	return newWASMDriver().GetName()
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
		wasm.mStateDB = state.NewMemoryStateDB(string(types.GetParaExec(tx.Execer)), wasm.GetStateDB(), wasm.GetLocalDB(), wasm.GetCoinsAccount(), wasm.GetHeight())
	}

	wasm.tx = tx
	wasm.txIndex = index
	//如果是调用wasm合约交易，则为期创建grpc调用channel
	if string(tx.Execer) != types.ExecName(wasmtypes.WasmX) {

		msgRecvOp := grpc.WithMaxMsgSize(wasmtypes.GRPCRecSize)
		if types.IsPara() && cfg.ParaRemoteGrpcClient == "" {
			panic("ParaRemoteGrpcClient error")
		}
		conn, err := grpc.Dial(cfg.ParaRemoteGrpcClient, grpc.WithInsecure(), msgRecvOp)
		if err != nil {
			panic(err)
		}
		wasm.grpcClient = types.NewChain33Client(conn)
	}
}

func (wasm *WASMExecutor) prepareQueryContext(executorName []byte) {
	if wasm.mStateDB == nil {
		wasm.mStateDB = state.NewMemoryStateDB(string(types.GetParaExec(executorName)), wasm.GetStateDB(), wasm.GetLocalDB(), wasm.GetCoinsAccount(), wasm.GetHeight())
	}
}

// 根据交易hash生成一个新的合约对象地址
//func (wasm *WASMExecutor) getNewAddr(txHash []byte) *address.Address {
//	return address.GetExecAddress(loccom.CalcWasmContractName(txHash))
//}

func (wasm *WASMExecutor) GenerateExecReceipt(usedGas, gasPrice uint64, snapshot int, execName, caller, contractAddr string, opType loccom.WasmContratOpType, debugInfo string) (*types.Receipt, error) {
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
	contractReceipt := &wasmtypes.ReceiptWASMContract{Caller:caller, ContractName:execName,ContractAddr: contractAddr, UsedGas:usedGas}

	runLog := &types.ReceiptLog{
		Ty:  wasmtypes.TyLogCallContractWasm,
		Log: types.Encode(contractReceipt)}
	if opType == wasmtypes.CreateWasmContractAction {
		runLog.Ty = wasmtypes.TyLogCreateUserWasmContract
	}

	//wasm子合约的debug信息
	debugLog := &wasmtypes.WasmDebugResp{DebugStatus: debugInfo}
	runLog2 := &types.ReceiptLog{
		Ty: wasmtypes.TyLogOutputItemWasm,
		Log: types.Encode(debugLog),
	}
	logs = append(logs, runLog2)

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
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))
	abi := wasm.mStateDB.GetAbi(contractAddr)
	if nil == abi {
		log.Error("getContractTable", "Failed to get abi for wasm contract", in.ContractName)
		return nil, wasmtypes.ErrAddrNotExists
	}
	wasm.mStateDB.SetCurrentExecutorName(string(types.GetParaExec([]byte(in.ContractName))))
	abi4CStr := C.CString(string(abi))
	defer C.free(unsafe.Pointer(abi4CStr))

	var wasmOutItems []*wasmtypes.WasmOutItem
	for _, item := range in.Items {
		data := wasm.mStateDB.GetState(contractAddr, item.Key)
		wasmOutItem := &wasmtypes.WasmOutItem{
			ItemType: item.TableName,
			Data:     data,
		}
		wasmOutItems = append(wasmOutItems, wasmOutItem)
	}

	for _, wasmOutItem := range wasmOutItems {
		if nil == wasmOutItem.Data {
			result := &wasmtypes.QueryResultItem{
				ItemType:   wasmOutItem.ItemType,
				ResultJSON: "Error:NO data saved in DB with such a key",
				Found:      false,
			}
			resp.QueryResultItems = append(resp.QueryResultItems, result)
			continue
		}
		var jsonResult *C.char
		structName := C.CString(wasmOutItem.ItemType)
		serializedData := C.CBytes(wasmOutItem.Data)

		log.Debug("wasm query", "structure", wasmOutItem.ItemType)

		if C.int(wasmtypes.Success) != C.convertData2Json(abi4CStr, (*C.char)(serializedData), (C.int)(len(wasmOutItem.Data)), structName, &jsonResult) {
			log.Error("wasm query", "structure", wasmOutItem.ItemType)

			result := &wasmtypes.QueryResultItem{
				ItemType:   wasmOutItem.ItemType,
				ResultJSON: "Error:Correct key value but with wrong Table name",
				Found:      false,
			}
			resp.QueryResultItems = append(resp.QueryResultItems, result)
			continue
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

type fuzzyDataItem struct {
	index int64
	Data  [][]byte
}

func (wasm *WASMExecutor) fuzzyGetContractTable(in *wasmtypes.WasmFuzzyQueryTableReq) (types.Message, error) {
	log.Debug("wasm fuzzy query", "WasmFuzzyQueryTableReq", in)

	resp := &wasmtypes.WasmFuzzyQueryResponse{
		ContractName:in.ContractName,
		TableName:in.TableName,
	}
	contractAddr := address.ExecAddress(types.ExecName(in.ContractName))
	wasm.prepareQueryContext([]byte(wasmtypes.WasmX))
	abi := wasm.mStateDB.GetAbi(contractAddr)
	if nil == abi {
		log.Error("getContractTable", "Failed to get abi for wasm contract", in.ContractName)
		return nil, wasmtypes.ErrAddrNotExists
	}
	wasm.mStateDB.SetCurrentExecutorName(string(types.GetParaExec([]byte(in.ContractName))))
	abi4CStr := C.CString(string(abi))
	defer C.free(unsafe.Pointer(abi4CStr))

	in.Format = string(types.LocalPrefix) + "-" + in.ContractName + "-data-" + contractAddr + ":" + in.Format

	var fuzzyDataItems []*fuzzyDataItem
	for i := in.Start; i <= in.Stop; i++  {
		prefix := []byte(fmt.Sprintf(in.Format, i))
		data := wasm.mStateDB.List(prefix)
		if nil == data {
			log.Debug("wasm_fuzzy_query_not_found", "Not found data for", string(prefix))
			continue
		}
		log.Debug("wasm_fuzzy_query_found_data", "Found data for", string(prefix))
		dataItem := &fuzzyDataItem{
			index:    i,
			Data:     data,
		}
		fuzzyDataItems = append(fuzzyDataItems, dataItem)
	}

	structName := C.CString(in.TableName)
	defer C.free(unsafe.Pointer(structName))
	for _, dataItemHeight := range fuzzyDataItems {
		fuzzyQueryResultItem := &wasmtypes.FuzzyQueryResultItem{
			Index:dataItemHeight.index,
		}
		for _, data := range dataItemHeight.Data {
			var jsonResult *C.char

			serializedData := C.CBytes(data)
			defer C.free(unsafe.Pointer(serializedData))

			if C.int(wasmtypes.Success) != C.convertData2Json(abi4CStr, (*C.char)(serializedData), (C.int)(len(data)), structName, &jsonResult) {
				log.Error("wasm fuzzy query convertData2Json failed", "structure", in.TableName)
				continue
			}

			fuzzyQueryResultItem.ResultJSON = append(fuzzyQueryResultItem.ResultJSON, C.GoString(jsonResult))
			C.free(unsafe.Pointer(jsonResult))

		}
		resp.QueryResultItems = append(resp.QueryResultItems, fuzzyQueryResultItem)
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

	exists := wasm.GetMStateDB().Exist(address.ExecAddress(types.ExecName(contractName)))
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
