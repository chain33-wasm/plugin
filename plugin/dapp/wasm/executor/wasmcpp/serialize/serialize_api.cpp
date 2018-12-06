/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#include <serialize_api.h>

using namespace std;

int unserializeAccount(char *data, int len, account4Go*pAccount) {
	return eosio::unserializeAccount(data, len, pAccount);
}

int unserializeCurrency_stats(char *data, int len, currency_stats4Go *pCurrency_stats) {
	return eosio::unserializeCurrency_stats(data, len, pCurrency_stats);
}

