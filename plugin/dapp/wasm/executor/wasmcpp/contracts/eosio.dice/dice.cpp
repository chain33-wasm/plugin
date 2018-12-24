#include "dice.hpp"
#define GAME_PRECISION 100
#define COIN_PRECISION 10000
#define STATUS "dice_statics"

namespace eosio {

using std::string;

void dice::startgame(int64_t deposit)
{ 
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
  string creator(fromBuf);

  string status_key(STATUS);
  //debug
  char debugInfo[512] = {0};
  sprintf(debugInfo, "Begin to startgame and creator:%s, deposit:%lld.\n"
  	"status_key:%s, with len:%d.\n", creator.c_str(), deposit,
  	status_key.c_str(), status_key.length());
  prints((const char *)debugInfo);
  //end debug
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  eosio_assert( valueSize == 0, "game already exists" );
  eosio_assert(deposit > 0, "deposit must be positive");
  eosio_assert(OK == execFrozenCoin(creator.c_str(), COIN_PRECISION * deposit), "fail to frozen coins");
  prints("Succeed to frozen coin\n");
  gamestatus status;
  status.height = getHeight4chain33();
  status.is_active = true;
  status.game_creator = creator;
  status.game_balance = GAME_PRECISION * deposit;
  status.current_round = 0;

  size_t size = pack_size( status );
  this->set_status(status);
}

//-x play -r "{\"player\":\"14cM9mnZ5JvbpFQTF4nwTCmix31VgmpweL\",\"amount\":\"2\",\"number\":\"40\"}" 
void dice::play(int64_t amount, uint8_t number, uint8_t direction)
{ 
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
  string player(fromBuf);
  gamestatus status = this->get_status();
  eosio_assert(status.is_active, "game is not active");
  eosio_assert(amount > 0, "amount must be positive");
  eosio_assert(number>=2 && number<= 97, "number must be within range of 2~97");
  eosio_assert(direction<=1, "direction must be 0 or 1");
  
  int64_t game_balance = status.game_balance;
  eosio_assert(GAME_PRECISION * 50 * amount < game_balance, "amount is too big");
  eosio_assert(OK == execFrozenCoin(player.c_str(), COIN_PRECISION * amount), "fail to frozen coins");
  int64_t probability = number;
    
  int64_t payout = GAME_PRECISION * amount * (100 - probability) / probability;
  printf("payout:%lld\n", payout);
  char arr[32] = {0};
  int length = get_random(arr, 32);
  eosio_assert(length >= 4, "get_random error");
  uint64_t a1 = uint64_t(arr[length - 1]);
  uint64_t a2 = uint64_t(arr[length - 2]) << 8;
  uint64_t a3 = uint64_t(arr[length - 3]) << 16;
  uint64_t a4 = uint64_t(arr[length - 4]) << 24;
  uint8_t rand_num = uint8_t((a1 + a2 + a3 +a4)%100);
  printf("rand num:%d\n",rand_num);

  roundinfo info;
  info.round = ++status.current_round;
  info.height = getHeight4chain33();
  info.player = player;
  info.amount = amount;
  info.guess_num = number;
  info.rand_num = rand_num;
  if (rand_num < number)
  {
    eosio_assert(OK == execTransferFrozenCoin(status.game_creator.c_str(), player.c_str(), COIN_PRECISION/GAME_PRECISION * payout), "fail to transfer frozen coins to player");
    status.game_balance -= payout;
    eosio_assert(OK == execActiveCoin(player.c_str(), COIN_PRECISION * amount), "fail to active coins of player");
    info.player_win = true;
    printf("you win\n");
  } 
  else 
  {
    eosio_assert(OK == execTransferFrozenCoin(player.c_str(), status.game_creator.c_str(), COIN_PRECISION * amount), "fail to transfer frozen coins to creator");
    eosio_assert(OK == execFrozenCoin(status.game_creator.c_str(), COIN_PRECISION * amount), "fail to frozen coins of creator");
	  status.game_balance += GAME_PRECISION * amount;
    info.player_win = false;
    printf("you lose\n");
  }
  this->add_roundinfo(info);
  status.height = info.height;
  this->set_status(status);
  this->set_localdb_for_height(info.height, info.round);
}

void dice::stopgame()
{
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
  string from(fromBuf);
  gamestatus status = this->get_status();
  string creator = status.game_creator;
  eosio_assert( from == creator, "game can only be stopped by creator" );
  eosio_assert(status.is_active, "game is not active");
  this->withdraw(creator);
  printf("withdraw\n");

  status.is_active = false;
  this->set_status(status);
  printf("stop game\n");

}

eosio::dice::gamestatus dice::get_status()
{
  string status_key(STATUS);
  int size = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  eosio_assert( size > 0, "Failed to get_status" );
  void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
  dbGet4chain33(status_key.c_str(), status_key.length(), (char *)buffer, size);
  datastream<char*> ds( (char*)buffer, size );
  gamestatus status;
  ds >> status;
  if (size > max_stack_buffer_size)
  {
    free(buffer);
  }
  return status;
}

void dice::set_status(gamestatus status)
{
  string status_key(STATUS);
  size_t size = pack_size(status);
  void* buffer = max_stack_buffer_size < size ? calloc(1, size) : alloca(size);
  datastream<char*> ds( (char*)buffer, size );
  ds << status;
  dbSet4chain33(status_key.c_str(), status_key.length(), (const char *)buffer, size);
  //debug
  prints_l(status_key.c_str(), status_key.length());
  char bufferDebug[256] = {0};
  sprintf(bufferDebug, "\nset_status key:%s with length:%d\n", status_key.c_str(), status_key.length());
  prints((const char*)bufferDebug);
  //end debug
  if (size > max_stack_buffer_size)
  {
    free(buffer);
  }
}

void dice::add_roundinfo(roundinfo info)
{
  char temp[64] = {0};
  sprintf(temp, "-height:%lld-round:%lld", info.height, info.round);
  string key(temp);
  size_t size = pack_size( info );
  void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
  datastream<char*> ds( (char*)buffer, size );
  ds << info;
  dbSet4chain33(key.c_str(), key.length(), (const char *)buffer, size);
		
  if (size > max_stack_buffer_size)
  {
    free(buffer);
  }
}

eosio::dice::heightinfo dice::get_localdb_for_height(string key)
{
  heightinfo info;
  info.start_round = 0;
  info.end_round = 0;
  int size = dbGetValueSize4chain33(key.c_str(), key.length());
  if (size > 0) {
    void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
    dbGet4chain33(key.c_str(), key.length(), (char *)buffer, size);
    datastream<char*> ds( (char*)buffer, size );
    ds >> info;
    if (size > max_stack_buffer_size)
    {
      free(buffer);
    }
  }

  return info;
}

void dice::set_localdb_for_height(int64_t height, int64_t round)
{
  char temp[64] = {0};
  sprintf(temp, "height:%lld", height);
  string key(temp);
  heightinfo info = this->get_localdb_for_height(key);
  if (info.start_round == 0) 
  {
    info.start_round = round;
  }
  info.end_round = round;

  size_t size = pack_size( info );
  void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
  datastream<char*> ds( (char*)buffer, size );
  ds << info;
  
  localdbSet4chain33(key.c_str(), key.length(), (const char *)buffer, size);	
  if (size > max_stack_buffer_size)
  {
    free(buffer);
  }
}

void dice::withdraw(string game_creator)
{
  int64_t balance = this->get_status().game_balance;
  eosio_assert(OK == execActiveCoin(game_creator.c_str(), balance/GAME_PRECISION*COIN_PRECISION), "fail to active coins");
}

bool dice::is_active()
{
  string status_key(STATUS);
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  if (valueSize == 0)
  {
    return false;
  }
  gamestatus status = this->get_status();
  return status.is_active;
}

}

EOSIO_ABI( eosio::dice, (startgame)(play)(stopgame))



