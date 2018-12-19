/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once

#include <eosiolib/asset.hpp>
#include <eosiolib/eosio.hpp>

#include <string>

namespace eosio {

   using std::string;

   class token : public contract {
      public:
         token( account_name self ):contract(self){}

         void create(string       issuer,
                            string       symbol,
                            int64_t      maximum_supply);

         void transfer(string       to,
	                     string       symbol,
                         int64_t      quantity,
                         string       memo );
		 void query(string owner, string symbol);
         // @abi table account i64
         typedef struct account {            
			int64_t 	 amount;
			string  symbol;           
         }account;
         // @abi table currency i64
         struct currency {
            string         symbol;
            int64_t        max_supply;
            string         issuer;
         };

      private:       

         void sub_balance(string owner, int64_t value, string symbol);
	 void add_balance(string owner, int64_t value, string symbol);

      public:
   };


} /// namespace eosio
