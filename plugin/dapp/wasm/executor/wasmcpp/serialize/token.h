/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once
#include <types.h>

// __cplusplus gets defined when a C++ compiler processes the file
#ifdef __cplusplus
// extern "C" is needed so the C++ compiler exports the symbols without name
// manging.
extern "C" {
#endif

typedef struct account4Go {            
	int64_t 	 amount;
	char  *symbol;
}account4Go;

typedef struct currency_stats4Go {
	char          *symbol;
	int64_t        max_supply;
	char         *issuer;
}currency_stats4Go;

int unserializeAccount(char *data, int len, account4Go *pAccount);

int unserializeCurrency_stats(char *data, int len, currency_stats4Go *pCurrency_stats);

#ifdef __cplusplus
}
#endif


