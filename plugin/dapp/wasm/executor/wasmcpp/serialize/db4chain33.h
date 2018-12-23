/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once

#include <eosiolib/types.h>
//#include <eosio/chain/_cgo_export.h>

#ifdef __cplusplus
extern "C" {
#endif

int get_valueSize(const void *key, int keyLen);
void pass_key(const void *key, int buffer_size);
void set_value(const void *value, int buffer_size);
int get_value(void *value, int buffer_size);
int64_t get_height();
int get_from(void *value, int buffer_size);

//The total accumlated size within one tx can't exceed 1M bytes,
//otherwise, the latter info will be ignored
void output2user(const char *type, const char* data, uint32_t len);

#ifdef __cplusplus
}
#endif

