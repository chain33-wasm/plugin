/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */

#include "token.hpp"

namespace eosio {
std::string genAccKey(string owner, string symbol);

void token::create( string       issuer,
                       string       symbol,
                       int64_t      maximum_supply)
{
    char temp[512] = {0};
	sprintf(temp, "token::create issuer:%s, symbol:%s, maximum_supply:%lld.\n", 
		issuer.c_str(), symbol.c_str(), maximum_supply);
	print((const char *)temp);
	
    auto sym = symbol;
    eosio_assert( maximum_supply > 0, "max-supply must be positive");
    
	sprintf(temp, "Cteate New token:%s", symbol.c_str());
	std::string key(temp);

	int valueSize = dbGetValueSize4chain33(key.c_str(), key.length());
    eosio_assert( valueSize == 0, "token with symbol already exists" );

	currency state;
	state.issuer        = issuer;
	state.symbol        = symbol;
	state.max_supply    = maximum_supply;

    constexpr size_t max_stack_buffer_size = 512;
	size_t size = pack_size( state );
	void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
    datastream<char*> ds( (char*)buffer, size );
    ds << state;
	
	dbSet4chain33(key.c_str(), key.length(), (const char *)buffer, size);

	account acc;
	acc.amount = maximum_supply;
	acc.symbol = symbol;
	
	std::string acckey = genAccKey(issuer, symbol);

	size = pack_size( acc );
	buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
    datastream<char*> ds2( (char*)buffer, size );
    ds2 << acc;
	dbSet4chain33(acckey.c_str(), acckey.length(), (const char *)buffer, size);
	
	//finishe token create operation
	sprintf(temp, "Finish token create operation.\n");
	print((const char *)temp);

	//
#if 0
	query(issuer, symbol);

	std::string to("12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv");
	transfer(to, symbol, 33, string("transfer 33"));
	query(to, symbol);
	query(issuer, symbol);
#endif
}

std::string genAccKey(string owner, string symbol) {
	char temp[128] = {0};
	sprintf(temp, "account info:%s-%s", symbol.c_str(), owner.c_str());
	std::string acckey(temp);
	print("New key genAccKey is:");
	print(acckey.c_str());
	print("\n");
	return acckey;
}

int getAccount(string owner, string symbol, token::account *pAccount) {
	string acckey = genAccKey(owner, symbol);
	char buffer[512];
	int size = dbGet4chain33(acckey.c_str(), acckey.length(), buffer, 512);
	if (size > 0) {
		datastream<char*> ds( (char*)buffer, size );	    
	    ds >> *pAccount;
		return 0;
	}
	
	return -1;
}

void updateAccount(string owner, string symbol, token::account accInfo) {
	string acckey = genAccKey(owner, symbol);	
	char buffer[512];
	int size = pack_size( accInfo );
	datastream<char*> ds( (char*)buffer, size );
	ds << accInfo;
	
	dbSet4chain33(acckey.c_str(), acckey.length(), buffer, size);	
}


void token::transfer(string       to,
	                     string       symbol,
                         int64_t      quantity,
                         string       memo )
{
    char temp[512] = {0};
	sprintf(temp, "On-going token::transfer to:%s, symbol:%s, quantity:%lld, memo:%s.\n", 
		to.c_str(), symbol.c_str(), quantity, memo.c_str());
	print((const char *)temp);

    char fromBuf[64] = {0};
    int fromsize = getFrom4chain33(fromBuf, 64);
	string from(fromBuf);    
    eosio_assert( from != to, "cannot transfer to self" );

    eosio_assert( quantity > 0, "must transfer positive quantity" );
    eosio_assert( memo.size() <= 256, "memo has more than 256 bytes" );


    print("\nOn-going sub_balance\n");
    sub_balance( from, quantity, symbol);
	print("\nOn-going add_balance\n");
    add_balance( to, quantity, symbol );
}

void token::sub_balance(string owner, int64_t value, string symbol) {
   token::account accinfo;
   getAccount(owner, symbol, &accinfo);
   eosio_assert( accinfo.amount > value, "owner's balance is not enough for transfer");
   accinfo.amount -= value;
   updateAccount(owner, symbol, accinfo);
}

void token::add_balance(string owner, int64_t value, string symbol) {
    token::account accinfo;
    int exist = getAccount(owner, symbol, &accinfo);
	if (exist == 0) {
		accinfo.amount += value;
		updateAccount(owner, symbol, accinfo);
	} else {
	    accinfo.symbol = symbol;
		accinfo.amount = value;
		updateAccount(owner, symbol, accinfo);
	}
   
}

void token::query(string owner, string symbol) {
	char temp[512] = {0};
	token::account accinfo;
    int exist = getAccount(owner, symbol, &accinfo);
	if (exist != 0) {
		sprintf(temp, " query_balance for owner:%s \n symbol:%s \n balance:0 \n", 
			owner.c_str(), symbol.c_str());
	} else {
	    sprintf(temp, " query_balance for owner:%s \n symbol:%s \n balance:%lld \n", 
			owner.c_str(), symbol.c_str(), accinfo.amount);
	}
	std::string info(temp);
	print(info.c_str());

	
	int size = pack_size( accinfo );
	datastream<char*> ds( (char*)temp, size );
	ds << accinfo;
	output2user("account", temp, size);
	print("Finish query with output2user\n");
}


} /// namespace eosio

EOSIO_ABI( eosio::token, (create)(transfer)(query))
