package types

import "errors"

var (
	ErrOutOfGasWASM                 = errors.New("out of gas for wasm")
	ErrCodeStoreOutOfGasWASM        = errors.New("contract creation code storage out of gas for wasm")
	ErrDepthWASM                    = errors.New("max call depth exceeded for wasm")
	ErrTraceLimitReachedWASM        = errors.New("the number of logs reached the specified limit for wasm")
	ErrInsufficientBalanceWASM      = errors.New("insufficient balance for transfer for wasm")
	ErrContractAddressCollisionWASM = errors.New("contract address collision for wasm")
	ErrGasLimitReachedWASM          = errors.New("gas limit reached for wasm")
	ErrGasUintOverflowWASM          = errors.New("gas uint64 overflow for wasm")
	ErrAddrNotExistsWASM            = errors.New("address not exists for wasm")
	ErrTransferBetweenContractsWASM = errors.New("transferring between contracts not supports for wasm")
	ErrTransferBetweenEOAWASM       = errors.New("transferring between external accounts not supports for wasm")
	ErrNoCreatorWASM                = errors.New("contract has no creator information for wasm")
	ErrDestructWASM                 = errors.New("contract has been destructed for wasm")

	ErrWriteProtectionWASM           = errors.New("wasm: write protection")
	ErrReturnDataOutOfBoundsWASM     = errors.New("wasm: return data out of bounds")
	ErrExecutionRevertedWASM         = errors.New("wasm: execution reverted")
	ErrMaxCodeSizeExceededWASM       = errors.New("wasm: max code size exceeded")
	ErrWrongContractAddr             = errors.New("wasm: wrong contract addr")
	ErrWASMValidationFail            = errors.New("wasm: fail to validate byte code")
	ErrWASMWavmNotSupported          = errors.New("wasm: vm wavm is not supported now")
	ErrWasmContractExecFailed        = errors.New("wasm: contract exec failed")
	ErrAddrNotExists                 = errors.New("wasm: contract addr not exists")
	ErrUnserialize                   = errors.New("wasm: unserialize")
	ErrCreateWasmPara                = errors.New("wasm: wrong parameter for creating new wasm contract")
	ErrCallWasmPara                  = errors.New("wasm: wrong parameter for creating call wasm contract")
)
