#include <string>
#include "dice.hpp"

using namespace std;
using eosio::dice;
using eosio::status_key;
using eosio::max_stack_buffer_size;

void dice::start_game(string creator, int64_t deposit)
{ 
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  eosio_assert( valueSize == 0, "game already exists" );
  eosio_assert(deposit > 0, "deposit must be positive");
  eosio_assert(OK == execFrozenCoin(creator.c_str(), deposit), "fail to frozen coins");
  gamestatus status;
  status.is_active = true;
  status.game_creator = creator;
  status.game_balance = deposit;
  status.current_round = 1;

	size_t size = pack_size( status );
  this->status_size = size;
  this->game_creator = creator;
  this->set_status(status);
}

void dice::play(string player, int64_t amount, int64_t number, int64_t direction)
{ 
  eosio_assert(this->is_active(), "game is not active");
  eosio_assert(amount > 0, "amount must be positive");
  eosio_assert(number>=0 && number<=99, "number must be within range of 0~99");
  eosio_assert(direction==0 || direction==1, "direction must be 0 or 1");
  
  int64_t game_balance = get_game_balance();
  eosio_assert(50 * amount < game_balance, "amount is too big");
  eosio_assert(OK == execFrozenCoin(player.c_str(), amount), "fail to frozen coins");
  int64_t probability = 0; //赢的概率
  if (direction == 0)
  {
    probability = number;
  }
  else
  {
    probability = 99-number;
  }
    
  int64_t payout = amount * (100 - probability) / probability;
  char temp[64] = {0};
  int length = GetRandom(temp, 64);
  eosio_assert(length==64, "GetRandom length error");
  int64_t rand_num = int64_t(temp[63]);
  printf("rand num:%lld\n",rand_num);
  roundinfo info;
  info.round = this->get_status_round() + 1;
  info.account = player;
  info.amount = amount;
  info.guess_num = number;
  info.result_num = rand_num;
  if ((direction==0&&rand_num<number) || (direction==1&&rand_num>number))
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(this->game_creator.c_str(), player.c_str(), payout), "fail to transfer frozen coins");
    this->change_game_balance(-payout);
    eosio_assert(OK == execActiveCoin(player.c_str(), amount+payout), "fail to active coins");
    info.player_win = true;
    printf("you win\n");
  }
  else
  {
    //TODO
    //保证原子操作？
    eosio_assert(OK == execTransferFrozenCoin(player.c_str(), this->game_creator.c_str(), amount), "fail to transfer frozen coins");
    this->change_game_balance(amount);
    info.player_win = false;
    printf("you lose\n");
  }
  this->add_status_round();
  this->add_roundinfo(info);
}

void dice::stop_game()
{
  char fromBuf[64] = {0};
  int fromsize = getFrom4chain33(fromBuf, 64);
	string from(fromBuf);
  gamestatus status = this->get_status();
  string creator = status.game_creator;
  eosio_assert( from == creator, "game can only be stopped by creator" );
  eosio_assert(this->is_active(), "game is not active");
  this->withdraw();
  printf("withdraw\n");

  status.is_active = false;
  this->set_status(status);
  this->status_size = 0;
  printf("stop game\n");

}

eosio::dice::gamestatus dice::get_status()
{
  size_t size = this->status_size;
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
  size_t size = this->status_size;
  void* buffer = max_stack_buffer_size < size ? malloc(size) : alloca(size);
  datastream<char*> ds( (char*)buffer, size );
  ds << status;
  dbSet4chain33(status_key.c_str(), status_key.length(), (const char *)buffer, size);
		
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
  sprintf(temp, "round:%lld,player:%s,amount:%lld,guess_num:%lld",
      info.round, info.account.c_str(), info.amount, info.guess_num);
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

void dice::withdraw()
{
  int64_t balance = this->get_game_balance();
  eosio_assert(OK == execActiveCoin(this->game_creator.c_str(), balance), "fail to active coins");
}

bool dice::is_active()
{
  int valueSize = dbGetValueSize4chain33(status_key.c_str(), status_key.length());
  if (valueSize == 0)
  {
    return false;
  }
  gamestatus status = this->get_status();
  return status.is_active;
}




