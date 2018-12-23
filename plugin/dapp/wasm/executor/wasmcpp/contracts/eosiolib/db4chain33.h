/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once

#include <eosiolib/types.h>

#ifdef __cplusplus
extern "C" {
#endif

int get_valueSize(const void *key, int keyLen);
void pass_key(const void *key, int buffer_size);
void set_value(const void *value, int buffer_size);
int get_value(void *value, int buffer_size);
int get_from(void *value, int buffer_size);
int get_random(char* randomDataOutput , int maxLen);
int get_LocalValueSize(const void *key, int keyLen);
void set_LocalValue(const void *value, int buffer_size);
int get_LocalValue(void *value, int buffer_size);

//The total accumlated size within one tx can't exceed 1M bytes,
//otherwise, the latter info will be ignored
void output2user(const char *type, const char* data, uint32_t len);
//////////interface for coin operation////////
#define Coin_Precision (10000)
//all the coin interface is operated with the 0.0001 precision,
//10000 denote 1, so 1 denote 0.0001
int execFrozenCoin(const char *addr, int64_t amount);
int execActiveCoin(const char *addr, int64_t amount);
int execTransferCoin(const char *from, const char *to, int64_t amount);
int execTransferFrozenCoin(const char *from, const char *to, int64_t amount);

//low precision interface
inline int execFrozenCoinLP(const char *addr, int64_t amount) {
	return execFrozenCoin(addr, Coin_Precision * amount);
}
inline int execActiveCoinLP(const char *addr, int64_t amount) {
	return execActiveCoin(addr, Coin_Precision * amount);
}
inline int execTransferCoinLP(const char *from, const char *to, int64_t amount) {
	return execTransferCoin(from, to, Coin_Precision * amount);
}
inline int execTransferFrozenCoinLP(const char *from, const char *to, int64_t amount) {
	return execTransferFrozenCoin(from, to, Coin_Precision * amount);
}


#ifdef __cplusplus
}
#endif

