#pragma once

#include <eosiolib/asset.hpp>
#include <eosiolib/eosio.hpp>
#include <string>

#define OK 0

namespace eosio {
    using std::string;
    constexpr size_t max_stack_buffer_size = 512;

    class dice : public contract{
    public:
        dice(account_name self):contract(self){};

        void startgame(int64_t deposit);
        void play(int64_t amount, uint8_t number, uint8_t direction);
        void stopgame();
        // @abi table roundinfo i64
        struct roundinfo {
            int64_t round;
            string player;
            int64_t amount;
            int64_t height;
            uint8_t guess_num;
            uint8_t rand_num;
            bool player_win;
        };
        // @abi table gamestatus i64
        struct gamestatus {
            bool is_active;
            string game_creator;
            int64_t height;
            int64_t game_balance;
            int64_t current_round;
        };

        // @abi table heightinfo i64
        struct heightinfo {
            int64_t start_round;
            int64_t end_round;
        };

    private:
        void withdraw(string game_creator);
        gamestatus get_status();
        void set_status(gamestatus status);
        void add_roundinfo(roundinfo round);
        heightinfo get_localdb_for_height(string key);
        void set_localdb_for_height(int64_t height, int64_t round);
        bool is_active();
    };
}
