package state

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"math/big"
)

type (
	// 检查制定账户是否有足够的金额进行转账
	CanTransferFunc func(StateDB, address.Address, address.Address, uint64) bool

	// 执行转账逻辑
	TransferFunc func(StateDB, address.Address, address.Address, uint64) bool

	// 获取制定高度区块的哈希
	// 给 BLOCKHASH 指令使用
	GetHashFunc func(uint64) common.Hash
)

type Context struct {
	// 下面这三个方法的说明，请查看方法类型的定义
	CanTransfer CanTransferFunc
	Transfer    TransferFunc
	GetHash     GetHashFunc

	// ORIGIN 指令返回数据， 合约调用者地址
	Origin *address.Address
	// GASPRICE 指令返回数据
	GasPrice uint32
	// GASLIMIT 指令，当前交易的GasLimit
	GasLimit uint64

	// NUMBER 指令，当前区块高度
	BlockNumber *big.Int
	// TIME 指令， 当前区块打包时间
	Time *big.Int
}