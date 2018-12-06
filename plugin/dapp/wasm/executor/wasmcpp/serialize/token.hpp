/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once
#include <string>

using namespace std; 
typedef struct account {            
	int64_t 	 amount;
	string  symbol;
}account;

typedef struct currency_stats {
	string         symbol;
	int64_t        max_supply;
	string         issuer;
}currency_stats;


int serializeAccount(char *data, int len, account& Account);	


