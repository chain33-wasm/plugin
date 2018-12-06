#include <eosiolib/eosio.hpp>
using namespace eosio;

class hello : public eosio::contract {
  public:
      using contract::contract;

      /// @abi action 
      void hi( account_name user ) {
         print( "Hello, ", name{user} );
		 
		 std::string key ("wasm contract key:Hello chain33.");
		 std::string value ("wasm contract value:Good luck!");
		 dbSet4chain33(key.c_str(), key.length(), value.c_str(), value.length());
      }
};

EOSIO_ABI( hello, (hi) )
