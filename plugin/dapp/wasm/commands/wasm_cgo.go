package commands

//#cgo CFLAGS: -I../../../external/abi/include/
//#cgo LDFLAGS: -L../../../external/abi/lib/ -labiconv -lboost_date_time -lstdc++
//#include <stdio.h>
//#include <stdlib.h>
//#include "abieos.h"
import "C"

import (
	"unsafe"
)

func genAbiData(contractAbi, contractName, actionName, abiJson string) []byte {
	contract := C.CString(contractName)
	defer C.free(unsafe.Pointer(contract))

	action := C.CString(actionName)
	defer C.free(unsafe.Pointer(action))

	abii := C.CString(contractAbi)
	defer C.free(unsafe.Pointer(abii))

	para := C.CString(abiJson)
	//para := C.CString("{\"user\":\"abcdf\"}")
	defer C.free(unsafe.Pointer(para))

	var abisize C.int
	abidata := C.genAbiFromJson(contract, action, abii, para, &abisize)
	defer C.free(unsafe.Pointer(abidata))

	abislice := C.GoBytes(unsafe.Pointer(abidata), abisize)
	return abislice
}
