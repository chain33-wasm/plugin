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


//The total accumlated size within one tx can't exceed 1M bytes,
//otherwise, the latter info will be ignored
void output2user(const char *type, const char* data, uint32_t len);
//////////interface for coin operation////////
int execFrozenCoin(const char *addr, int64_t amount);
int execActiveCoin(const char *addr, int64_t amount);
int execTransferCoin(const char *from, const char *to, int64_t amount);
int execTransferFrozenCoin(const char *from, const char *to, int64_t amount);


#ifdef __cplusplus
}
#endif

