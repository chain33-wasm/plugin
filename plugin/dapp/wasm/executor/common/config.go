package common

const (
	// 最大Gas消耗上限
	MaxGasLimit = 10000000
	MaxCodeSize = 2*1024*1024 // 合约允许的最大字节数,待定

	// 各种操作对应的Gas定价
	CreateDataGas        uint64 = 200   // 创建合约时，按字节计费
)

// 解释器的配置模型
type Config struct {
}
