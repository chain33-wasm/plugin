/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once

#include <eosiolib/datastream.h>

namespace eosio {
 
typedef struct account {            
	int64_t 	 amount;
	string  symbol;
}account;

typedef struct currency_stats {
	string         symbol;
	int64_t        max_supply;
	string         issuer;
}currency_stats;

int serializeAccount(char *data, int len, account& Account) {
	int size = pack_size( Account );
	if (len >= size) {
		datastream<char*> ds( (char*)data, size );	    
		ds << Account;

		return size;
	}
	
	return 0;
}


int unserializeAccount(char *data, int len, account *pAccount) {
	char buffer[512];
	memcpy(buffer, data, len);
	datastream<char*> ds( (char*)buffer, len );	    
	ds >> *pAccount;

	return 0;
}

int unserializeCurrency_stats(char *data, int len, currency_stats *pCurrency_stats) {
	char buffer[512];
	memcpy(buffer, data, len);
	datastream<char*> ds( (char*)buffer, len );	    
	ds >> *pCurrency_stats;

	return 0;
}


} /// namespace eosio
