
#include "dice.hpp"

namespace eosio {

using namespace std;
using eosio::dice;
using eosio::max_stack_buffer_size;

void dice::startgame(string creator, int64_t deposit)
{ 
  string status_key("creator status");
  char debugInfo[512] = {0};
  sprintf(debugInfo, "Begin to startgame and creator:%s, deposit:%lld.\n"
  	"status_key:%s, with len:%d.\n", creator.c_str(), deposit,
  	status_key.c_str(), status_key.length());
  prints((const char *)debugInfo);
  
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  eosio_assert( valueSize == 0, "game already exists" );
  eosio_assert(deposit > 0, "deposit must be positive");
  eosio_assert(OK == execFrozenCoin(creator.c_str(), deposit), "fail to frozen coins");
  prints("Succeed to frozen coin\n");
  gamestatus status;
  status.is_active = true;
  status.game_creator = creator;
  status.game_balance = deposit;
  status.current_round = 0;

  size_t size = pack_size( status );
  this->set_status(status);
}

//-x play -r "{\"player\":\"14cM9mnZ5JvbpFQTF4nwTCmix31VgmpweL\",\"amount\":\"2\",\"number\":\"40\",\"direction\":\"0\"}" 
void dice::play(string player, int64_t amount, int64_t number, int64_t direction)
{ 
  eosio_assert(this->is_active(), "game is not active");
  eosio_assert(amount > 0, "amount must be positive");
  eosio_assert(number>=2 && number<= 97, "number must be within range of 2~97");
  eosio_assert(direction==0 || direction==1, "direction must be 0 or 1");
  
  int64_t game_balance = get_game_balance();
  eosio_assert(50 * amount < game_balance, "amount is too big");
  eosio_assert(OK == execFrozenCoin(player.c_str(), amount), "fail to frozen coins");
  int64_t probability = 0; //赢的概率
  //guess small
  if (direction == 0)
  {
    probability = number;
  }
  //guess big
  else
  {
    probability = 100 - number;
  }
    
  int64_t payout = amount * (100 - probability) / probability;
  char temp[32] = {0};
  int length = get_random(temp, 32);
  eosio_assert(length > 0, "get_random error");
  int64_t rand_num = int64_t(temp[length - 1]);
  rand_num = (rand_num * 100) >> 4; // *100/16
  print_f("rand num:%lld\n",rand_num);
  roundinfo info;
  info.round = this->get_status_round() + 1;
  info.account = player;
  info.amount = amount;
  info.guess_num = number;
  info.result_num = rand_num;
  gamestatus status = this->get_status();
  if ((direction==0 && rand_num < number) || (direction==1 && rand_num > number))
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(status.game_creator.c_str(), player.c_str(), payout), "fail to transfer frozen coins");
    this->change_game_balance(-payout);
    eosio_assert(OK == execActiveCoin(player.c_str(), amount), "fail to active coins");
    info.player_win = true;
    printf("you win\n");
  }
  else
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(player.c_str(), status.game_creator.c_str(), amount), "fail to transfer frozen coins");
    eosio_assert(OK == execFrozenCoin(status.game_creator.c_str(), amount), "fail to frozen coins");
	this->change_game_balance(amount);
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
  eosio_assert(this->is_active(), "game is not active");
  this->withdraw(creator);
  printf("withdraw\n");

  status.is_active = false;
  this->set_status(status);
  printf("stop game\n");

}

eosio::dice::gamestatus dice::get_status()
{
  string status_key("creator status");
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
  string status_key("creator status");
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
  eosio_assert(OK == execActiveCoin(game_creator.c_str(), balance), "fail to active coins");
}

bool dice::is_active()
{
  string status_key("creator status");
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



