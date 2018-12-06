/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */

#include <datastream.hpp>
#include <token.hpp>
#include <token.h>

int serializeAccount(char *data, int len, account& Account) {
	int size = eosio::pack_size( Account );
	if (len >= size) {
		eosio::datastream<char*> ds( (char*)data, size );	    
		ds << Account;

		return size;
	}
	
	return 0;
}

int unserializeAccount(char *data, int len, account4Go *pAccount) {
	char buffer[512];
	memcpy(buffer, data, len);
	eosio::datastream<char*> ds( (char*)buffer, len );
	account acc;
	ds >> acc;

	pAccount->amount = acc.amount;
	char *strINfo = (char *)malloc(acc.symbol.size());
	memcpy(strINfo, acc.symbol.c_str(), acc.symbol.size());
	pAccount->symbol = strINfo;		

	return 0;
}

int unserializeCurrency_stats(char *data, int len, currency_stats4Go *pCurrency_stats) {
	char buffer[512];
	memcpy(buffer, data, len);
	eosio::datastream<char*> ds( (char*)buffer, len );
	currency_stats states;
	ds >> states;

	char *strINfo = (char *)malloc(states.symbol.size());
	memcpy(strINfo, states.symbol.c_str(), states.symbol.size());
	pCurrency_stats->symbol = strINfo;

    strINfo = (char *)malloc(states.issuer.size());
	memcpy(strINfo, states.issuer.c_str(), states.issuer.size());
	pCurrency_stats->issuer = strINfo;
	
	pCurrency_stats->max_supply = states.max_supply;
	
	return 0;
}

