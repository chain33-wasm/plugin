#pragma once

#include <eosiolib/asset.hpp>
#include <eosiolib/eosio.hpp>
#include <string>

#define OK 0

namespace eosio {
    using std::string;
    const string status_key = "creator status";
    constexpr size_t max_stack_buffer_size = 512;

    class dice : public contract{
    public:
        dice(account_name self):contract(self){};

        void startgame(string creator, int64_t deposit);
        void play(string player, int64_t amount, int64_t number, int64_t direction);
        void stopgame();
        // @abi table roundinfo i64
        struct roundinfo {
            int64_t round;
            string account;
            int64_t amount;
            int64_t guess_num;
            int64_t result_num;
            bool player_win;
        };
        // @abi table gamestatus i64
        struct gamestatus {
            bool is_active;
            string game_creator;
            //string game_addr;
            int64_t game_balance;
            int64_t current_round;
        };

    private:
        size_t status_size = 0;
        string game_creator;
        void withdraw();
        gamestatus get_status();
        void set_status(gamestatus status);
        int64_t get_game_balance();
        void change_game_balance(int64_t amount);
        void add_roundinfo(roundinfo round);
        int64_t get_status_round();
        void add_status_round();
        bool is_active();
    };
}

int execFrozenCoin(const char *addr, int64_t amount);
int execActiveCoin(const char *addr, int64_t amount);
int execTransferCoin(const char *from, const char *to, int64_t amount);
int execTransferFrozenCoin(const char *from, const char *to, int64_t amount);