#include <eosiolib/eosio.hpp>
using namespace eosio;

class hello : public eosio::contract {
  public:
      using contract::contract;

      /// @abi action 
      void hi( account_name user ) {
         print( "Contract DB write and read for Hello, ", name{user}.to_string().c_str() );
		 print("\nWelcome to use our wasm contract platform!", "It's developed by hezhengjun of Chain33.\n");
		 
		 std::string key ("wasm contract key:Hello chain33.");
		 std::string value ("wasm contract value:Good luck!");
		 dbSet4chain33(key.c_str(), key.length() + 1, value.c_str(), value.length() + 1);
		 int valSize = dbGetValueSize4chain33(key.c_str(), key.length());
		 if (valSize > 0) {
		 	 constexpr size_t max_stack_buffer_size = 512;
             void* buffer = nullptr;
             buffer = max_stack_buffer_size < valSize ? malloc(valSize) : alloca(valSize);
		     int getsize = dbGet4chain33(key.c_str(), key.length() + 1, (char *)buffer, valSize);

			 std::string temp("Contract DB retrieve value is: ");
		     temp += (char *)buffer;
			 print(temp);
			 if ( max_stack_buffer_size < valSize ) {
                free(buffer);
            }
		 }
		  
      }
};

EOSIO_ABI( hello, (hi) )
