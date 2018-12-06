#pragma once

#include <eosio/chain/name.hpp>
#include <eosio/chain/chain_id_type.hpp>
#include <fc/string.hpp>

#include <vector>
#include "../../../../chainbase/include/chainbase/chainbase.hpp"




namespace eosio { namespace chain {

   using                     std::vector;
   using                     std::string;
   using                     std::unique_ptr;

   using chainbase::allocator;
   using shared_string = boost::interprocess::basic_string<char, std::char_traits<char>, allocator<char>>;
   
   using checksum_type		 = fc::sha256;
   using digest_type 		 = checksum_type;   
   using bytes               = vector<char>;


} }  // eosio::chain


