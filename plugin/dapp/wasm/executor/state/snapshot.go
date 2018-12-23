package state

import (
	"sort"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	loccom "github.com/33cn/plugin/plugin/dapp/wasm/executor/common"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

// 数据状态变更接口
// 所有的数据状态变更事件实现此接口，并且封装各自的变更数据以及回滚动作
// 在调用合约时（具体的Tx执行时），会根据操作生成对应的变更对象并缓存下来
// 如果合约执行出错，会按生成顺序的倒序，依次调用变更对象的回滚接口进行数据回滚，并同步删除变更对象缓存
// 如果合约执行成功，会按生成顺序的郑旭，依次调用变更对象的数据和日志变更记录，回传给区块链
type DataChange interface {
	revert(mdb *MemoryStateDB)
	getData(mdb *MemoryStateDB) []*types.KeyValue
	getLog(mdb *MemoryStateDB) []*types.ReceiptLog
}

// 版本结构，包含版本号以及当前版本包含的变更对象在变更序列中的开始序号
type Snapshot struct {
	id      int
	entries []DataChange
	statedb *MemoryStateDB
}

func (ver *Snapshot) GetId() int {
	return ver.id
}

// 回滚当前版本
func (ver *Snapshot) revert() bool {
	if ver.entries == nil {
		return true
	}
	for _, entry := range ver.entries {
		entry.revert(ver.statedb)
	}
	return true
}

// 添加变更数据
func (ver *Snapshot) append(entry DataChange) {
	ver.entries = append(ver.entries, entry)
}

// 获取当前版本变更数据
func (ver *Snapshot) getData(opType loccom.WasmContratOpType) (kvSet []*types.KeyValue, logs []*types.ReceiptLog) {
	// 获取中间的数据变更
	dataMap := make(map[string]*types.KeyValue)

	//var localKvSet []*types.KeyValue
	//localDataMap := make(map[string]*types.KeyValue)
	//var wasmContractName string
	for _, entry := range ver.entries {

		items := entry.getData(ver.statedb)
		logEntry := entry.getLog(ver.statedb)
		if logEntry != nil {
			logs = append(logs, entry.getLog(ver.statedb)...)
		}
		////当前的storage的状态变化仅存放在localDB中，所有需要将其进行汇总，作为第二步的备用数据
		//if wasmtypes.CallWasmContractAction == opType {
		//	if stChg, ok := entry.(storageChange); ok {
		//		wasmContractName = ver.statedb.GetAccount(stChg.account).Data.Name
		//		localKvSet = append(localKvSet, stChg.getDataFromLocalDB(ver.statedb)...)
		//	}
		//}

		// 执行去重操作
		for _, kv := range items {
			dataMap[string(kv.Key)] = kv
		}
	}

	////////////////////////////
	//因为调用合约时的stroagechange的数据存储在localdb中，为保证数据的一致性，需要将
	//将所有storagechange的数据key和value进行append，进行hash计算并进行存储
	//if wasmtypes.CallWasmContractAction == opType {
	//	for _, kv := range localKvSet {
	//		localDataMap[string(kv.Key)] = kv
	//	}
	//	locnames := make([]string, 0, len(localDataMap))
	//	for name := range localDataMap {
	//		locnames = append(locnames, name)
	//	}
	//	sort.Strings(locnames)
	//
	//	var keys []byte
	//	var values []byte
	//	for _, name := range locnames {
	//		keys = append(keys, localDataMap[name].Key...)
	//		values = append(values, localDataMap[name].Value...)
	//	}
	//	keyHash := common.Sha256(keys)
	//	valueHash := common.Sha256(values)
	//	key := loccom.CalcStrorageChangeKey(wasmContractName, common.ToHex(keyHash))
	//	kvSet = append(kvSet, &types.KeyValue{Key: key, Value: valueHash})
	//}
	///////////////////////////////////

	// 这里也可能会引起数据顺序不一致的问题，需要修改（目前看KV的顺序不会影响哈希计算，但代码最好保证顺序一致）
	names := make([]string, 0, len(dataMap))
	for name := range dataMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		kvSet = append(kvSet, dataMap[name])
	}

	return kvSet, logs
}

type (

	// 基础变更对象，用于封装默认操作
	baseChange struct {
	}

	// 创建合约对象变更事件
	createAccountChange struct {
		baseChange
		account string
	}

	// 自杀事件
	suicideChange struct {
		baseChange
		account string
		prev    bool // whether account had already suicided
	}

	// nonce变更事件
	nonceChange struct {
		baseChange
		account string
		prev    uint64
	}

	// 存储状态变更事件
	storageChange struct {
		baseChange
		account  string
		key      []byte
		prevalue []byte
	}

	// 本地存储状态变更事件
	localStorageChange struct {
		baseChange
		account  string
		key      []byte
		data     []byte
		prevalue []byte
	}

	// 合约代码状态变更事件
	codeChange struct {
		baseChange
		account  string
		prevcode []byte
		prevhash []byte
		prevabi  []byte
	}

	// 返还金额变更事件
	refundChange struct {
		baseChange
		prev uint64
	}

	// 转账事件
	// 合约转账动作不执行回滚，失败后数据不会写入区块
	balanceChange struct {
		baseChange
		amount int64
		data   []*types.KeyValue
		logs   []*types.ReceiptLog
	}

	// 合约生成日志事件
	addLogChange struct {
		baseChange
		txhash common.Hash
	}

	// 合约生成sha3事件
	addPreimageChange struct {
		baseChange
		hash common.Hash
	}
)

// 在baseChang中定义三个基本操作，子对象中只需要实现必要的操作
func (ch baseChange) revert(s *MemoryStateDB) {
}

func (ch baseChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	return nil
}

func (ch baseChange) getLog(s *MemoryStateDB) (logs []*types.ReceiptLog) {
	return nil
}

// 创建账户对象的回滚，需要删除缓存中的账户和变更标记
func (ch createAccountChange) revert(s *MemoryStateDB) {
	delete(s.accounts, ch.account)
}

// 创建账户对象的数据集
func (ch createAccountChange) getData(s *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := s.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		kvset = append(kvset, acc.GetStateKV()...)
		return kvset
	}
	return nil
}

func (ch suicideChange) revert(mdb *MemoryStateDB) {
	// 如果已经自杀过了，不处理
	if ch.prev {
		return
	}
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.State.Suicided = ch.prev
	}
}

func (ch suicideChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	// 如果已经自杀过了，不处理
	if ch.prev {
		return nil
	}
	acc := mdb.accounts[ch.account]
	if acc != nil {
		return acc.GetStateKV()
	}
	return nil
}

func (ch nonceChange) revert(mdb *MemoryStateDB) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.State.Nonce = ch.prev
	}
}

func (ch nonceChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	// nonce目前没有应用场景，而且每次调用都会变更，暂时先不写到状态数据库中
	//acc := mdb.accounts[ch.account]
	//if acc != nil {
	//	return acc.GetStateKV()
	//}
	return nil
}

func (ch codeChange) revert(mdb *MemoryStateDB) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.Data.Code = ch.prevcode
		acc.Data.CodeHash = ch.prevhash
	}
}

func (ch codeChange) getData(mdb *MemoryStateDB) (kvset []*types.KeyValue) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		kvset = append(kvset, acc.GetDataKV()...)
		kvset = append(kvset, acc.GetStateKV()...)
		return kvset
	}
	return nil
}

func (ch storageChange) revert(mdb *MemoryStateDB) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		acc.SetState(string(ch.key), ch.prevalue)
	}
}

func (ch storageChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	value := mdb.GetState(ch.account, string(ch.key))
	if value == nil {
		return nil
	}
	acc := mdb.GetAccount(ch.account)
	key := acc.GetStateItemKey(ch.account, string(ch.key))

	return []*types.KeyValue{{Key: []byte(key), Value: value}}
}

func (ch storageChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	return nil
}

func (ch storageChange) getDataFromLocalDB(mdb *MemoryStateDB) []*types.KeyValue {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		currentVal := acc.GetState(string(ch.key))
		var kvSet []*types.KeyValue
		kvSet = append(kvSet, &types.KeyValue{Key: []byte(ch.key), Value: currentVal})
		return kvSet
	}
	return nil
}

func (ch refundChange) revert(mdb *MemoryStateDB) {
	mdb.refund = ch.prev
}

func (ch addLogChange) revert(mdb *MemoryStateDB) {
	logs := mdb.logs[ch.txhash]
	if len(logs) == 1 {
		delete(mdb.logs, ch.txhash)
	} else {
		mdb.logs[ch.txhash] = logs[:len(logs)-1]
	}
	mdb.logSize--
}

func (ch addPreimageChange) revert(mdb *MemoryStateDB) {
	delete(mdb.preimages, ch.hash)
}

func (ch balanceChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	return ch.data
}
func (ch balanceChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	return ch.logs
}

func (ch localStorageChange) revert(mdb *MemoryStateDB) {
	acc := mdb.accounts[ch.account]
	if acc != nil {
		mdb.LocalDB.Set(ch.key, ch.prevalue)
	}
}

func (ch localStorageChange) getData(mdb *MemoryStateDB) []*types.KeyValue {
	return nil
}

func (ch localStorageChange) getLog(mdb *MemoryStateDB) []*types.ReceiptLog {
	localData := &wasmtypes.ReceiptLocalData{
		Key:ch.key,
		CurValue:ch.data,
		PreValue:ch.prevalue,
	}

	log := &types.ReceiptLog{
		Ty:wasmtypes.TyLogLocalDataWasm,
		Log:types.Encode(localData),
	}

	return []*types.ReceiptLog{log}
}
