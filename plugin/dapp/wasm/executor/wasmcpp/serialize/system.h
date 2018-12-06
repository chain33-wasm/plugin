/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once
#include <types.h>
#include <fc/exception/exception.hpp>
#include <eosio/chain/exceptions.hpp>

extern "C" {

void eosio_assert( bool condition, const char *msg );
}
