/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once
#include <eosiolib/db4chain33.h>
#include <eosiolib/types.hpp>
#include <eosiolib/fixed_key.hpp>
#include <utility>
#include <string>


namespace eosio {

inline int dbGet4chain33(const char *key, int keyLen, char *pvalue, int vallen) {	
	pass_key(key, keyLen);
	return get_value(pvalue, vallen);
}

inline void dbSet4chain33(const char *key, int keyLen, const char *pvalue, int vallen) {
    pass_key(key, keyLen);
	set_value(pvalue, vallen);
}

inline int dbGetValueSize4chain33(const char *key, int keyLen) {	
    return get_valueSize(key, keyLen);
}

inline int getFrom4chain33(char *from, int fromLen) {	
    return get_from(from, fromLen);
}

inline int64_t getHeight4chain33() {
    return get_height();
}
}

