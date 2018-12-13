/* Created by "go tool cgo" - DO NOT EDIT. */

/* package executor */

/* Start of preamble from import "C" comments.  */


#line 3 "/home/hezhengjun/work/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasm.go"




#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../../../../../wasmcpp/adapter/include/eosio/chain/wasm_interface_adapter.h"

#line 1 "cgo-generated-wrapper"


/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


//在获取key对应的value之前，需要先获取下value的size，为了避免传递的内存太小
extern int StateDBGetValueSizeCallback (char* p0, char* p1, int p2);
extern int StateDBGetStateCallback (char* p0, char* p1, int p2, char* p3, int p4);
extern void StateDBSetStateCallback (char* p0, char* p1, int p2, char* p3, int p4);

//该接口用于返回查询结果的返回
extern void Output2UserCallback (char* p0, char* p1, int p2);

////////////以下接口用于user.wasm.xxx合约内部转账/////////////////////////////
//冻结user.wasm.xxx合约addr上的部分余额,其中的
extern int ExecFrozen (char* p0, long long int p1);

//激活user.wasm.xxx合约addr上的部分余额
extern int ExecActive (char* p0, long long int p1);
extern int ExecTransfer (char* p0, char* p1, long long int p2);
extern int ExecTransferFrozen (char* p0, char* p1, long long int p2);

//为wasm用户自定义合约提供随机数，该随机数是64位hash值,返回值为实际获取到的长度
extern int func GetRandom(char* randomDataOutput , int maxLen);
#ifdef __cplusplus
}
#endif
