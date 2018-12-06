#include <system.h>
#include <fc/exception/exception.hpp>
#include <eosio/chain/exceptions.hpp>


   void eosio_assert( bool condition, const char *msg ) {
      if( BOOST_UNLIKELY( !condition ) ) {
         std::string message( msg );
         edump((message));
         EOS_THROW( eosio::chain::eosio_assert_message_exception, "assertion failure with message: ${s}", ("s",message) );
      }
   }
