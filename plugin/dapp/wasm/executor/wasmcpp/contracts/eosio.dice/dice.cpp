#include "dice.hpp"
#define GAME_PRECISION 100
#define COIN_PRECISION 10000
#define STATUS "creator status"

namespace eosio {

using namespace std;

void dice::startgame(int64_t deposit)
{ 
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
  string creator(fromBuf);

  string status_key(STATUS);
  char debugInfo[512] = {0};
  sprintf(debugInfo, "Begin to startgame and creator:%s, deposit:%lld.\n"
  	"status_key:%s, with len:%d.\n", creator.c_str(), deposit,
  	status_key.c_str(), status_key.length());
  prints((const char *)debugInfo);
  
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  eosio_assert( valueSize == 0, "game already exists" );
  eosio_assert(deposit > 0, "deposit must be positive");
  eosio_assert(OK == execFrozenCoin(creator.c_str(), COIN_PRECISION * deposit), "fail to frozen coins");
  prints("Succeed to frozen coin\n");
  gamestatus status;
  status.is_active = true;
  status.game_creator = creator;
  status.game_balance = GAME_PRECISION * deposit;
  status.current_round = 0;

  size_t size = pack_size( status );
  this->set_status(status);
}

//-x play -r "{\"player\":\"14cM9mnZ5JvbpFQTF4nwTCmix31VgmpweL\",\"amount\":\"2\",\"number\":\"40\",\"direction\":\"0\"}" 
void dice::play(int64_t amount, int64_t number)
{ 
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
  string player(fromBuf);
  gamestatus status = this->get_status();
  eosio_assert(status.is_active, "game is not active");
  eosio_assert(amount > 0, "amount must be positive");
  eosio_assert(number>=2 && number<= 97, "number must be within range of 2~97");
  
  int64_t game_balance = get_game_balance();
  eosio_assert(GAME_PRECISION * 50 * amount < game_balance, "amount is too big");
  eosio_assert(OK == execFrozenCoin(player.c_str(), COIN_PRECISION * amount), "fail to frozen coins");
  int64_t probability = number;
    
  int64_t payout = GAME_PRECISION * amount * (100 - probability) / probability;
  char arr[32] = {0};
  int length = get_random(arr, 32);
  eosio_assert(length > 0, "get_random error");
  char temp = arr[length - 1] & 0x0f;
  int64_t rand_num = (int64_t(temp) * 100) >> 4; // *100/16
  print_f("rand num:%lld\n",rand_num);
  roundinfo info;
  info.round = this->get_status_round() + 1;
  info.account = player;
  info.amount = amount;
  info.guess_num = number;
  info.result_num = rand_num;
  if (rand_num < number)
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(status.game_creator.c_str(), player.c_str(), COIN_PRECISION * payout), "fail to transfer frozen coins");
    this->change_game_balance(- GAME_PRECISION * payout);
    eosio_assert(OK == execActiveCoin(player.c_str(), COIN_PRECISION * amount), "fail to active coins");
    info.player_win = true;
    printf("you win\n");
  }
  else
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(player.c_str(), status.game_creator.c_str(), COIN_PRECISION * amount), "fail to transfer frozen coins");
    eosio_assert(OK == execFrozenCoin(status.game_creator.c_str(), COIN_PRECISION * amount), "fail to frozen coins");
	this->change_game_balance(GAME_PRECISION * amount);
    info.player_win = false;
    printf("you lose\n");
  }
  this->add_status_round();
  this->add_roundinfo(info);
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
  eosio::dbGet4chain33(status_key.c_str(), status_key.length(), (char *)buffer, size);
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
  //debug/
  prints_l(status_key.c_str(), status_key.length());
  char bufferDebug[256] = {0};
  sprintf(bufferDebug, "\nset_status key:%s with length:%d\n", status_key.c_str(), status_key.length());
  prints((const char*)bufferDebug);
  ///////////////
		
  if (size > max_stack_buffer_size)
  {
    free(buffer);
  }
}

int64_t dice::get_game_balance()
{
  gamestatus status = this->get_status();
  return status.game_balance;
}

void dice::change_game_balance(int64_t change)
{
  gamestatus status = this->get_status();
  status.game_balance += change;
  this->set_status(status);
}

void dice::add_roundinfo(roundinfo info)
{
  char temp[64] = {0};
  sprintf(temp, "round:%lld", info.round);
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

int64_t dice::get_status_round()
{
  gamestatus status = this->get_status();
  return  status.current_round;
}

void dice::add_status_round()
{
  gamestatus status = this->get_status();
  status.current_round++;
  this->set_status(status);
}

void dice::withdraw(string game_creator)
{
  int64_t balance = this->get_game_balance();
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



