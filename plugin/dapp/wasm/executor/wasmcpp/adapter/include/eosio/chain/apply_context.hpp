/**
 *  @file
 *  @copyright defined in eos/LICENSE.txt
 */
#pragma once
#include <eosio/chain/action.hpp>
#include <eosio/chain/wasm_interface_adapter.h>
#include <fc/utility.hpp>
#include <sstream>
#include <algorithm>
#include <set>

namespace eosio { namespace chain {

#define GAS_CHECK_FAIL    (-1L)
#define GAS_CHECK_SUCCEES (0L)
#define MAX_OUT2USER_SIZE (0x100000L)
typedef int gas_check_res;

extern uint64_t SstoreSetGas; 
extern uint64_t SstoreResetGas;  
extern uint64_t SstoreClearGas;
extern uint64_t SstoreLoadGas;

typedef struct {
    string       type;
	vector<char> data;
} Out2UserElement;


class apply_context {

   /// Constructor
   public:
      apply_context(Apply_context_para *pApply_context, uint32_t depth=0)
      :contractAddr(std::string(pApply_context->contractAddr)),
      gasAvailable(pApply_context->gasAvailable),
      currentBlockTime(pApply_context->blocktime),
      from(std::string(pApply_context->from)),
      height(pApply_context->height),
      recurse_depth(depth)
      {
         bytes data(pApply_context->datalen);
	     for(bytes::iterator iter = data.begin(); iter != data.end(); iter++) {
             *iter = *pApply_context->pdata++;
	     }
	     action act(string_to_name(&pApply_context->contractName[0]), string_to_name(&pApply_context->action_name[0]), data);
	     actInfo = act;
		 receiver = actInfo.account;
		 
         std::cout<<"apply_context is created now and act data size is:"<<actInfo.data.size()<<"\n";
		 std::cout<<"contractAddr is:"<<contractAddr<<"\n";
         reset_console();
      }
	virtual ~apply_context();
	void reset_console();
	std::ostringstream& get_console_stream()            { return _pending_console_output; }
	const std::ostringstream& get_console_stream()const { return _pending_console_output; }

	template<typename T>
	void console_append(T val) {
         _pending_console_output << val;
      }

      template<typename T, typename ...Ts>
      void console_append(T val, Ts ...rest) {
         console_append(val);
         console_append(rest...);
      };

      inline void console_append_formatted(const string& fmt, const variant_object& vo) {
         console_append(fc::format_string(fmt, vo));
      }

	  gas_check_res check_and_spend_gas(int gascost);
	  void setKey(const char *data, int len);
	  void setValue(const char *data, int len);
	  int getValue(char *data, int len);	
	  int getValueSize(const char *data, int len);
	  int64_t getBlockTime()const;
	  int get_from(char *from, size_t from_size);
	  void output2user(const char *type, const char *data, int len);
	  int execFrozenCoin(char* addr, long long int p1);
	  int execActiveCoin(char* addr, long long int p1);
	  int execTransferCoin(char* from, char* to, long long int p2);
	  int execTransferFrozenCoin(char* from, char* to, long long int p2);
	  
   /// Fields:  
      action                  actInfo; ///< message being applied
      ///< the code that is currently running, actully, it's contract code-id
      account_name                  receiver;
	  string                        contractAddr;
      uint32_t                      recurse_depth; ///< how deep inline actions can recurse
      bool                          context_free = false;
      bool                          used_context_free_api = false;
      //action_trace                  trace;
      std::ostringstream            _pending_console_output;
	  int64_t                       gasAvailable;
	  
    private:
	  vector<char>                  key4DB;
	  vector<char>                  value4DB;
	  vector<Out2UserElement>       out2user;
	  int64_t                       currentBlockTime;
	  int64_t                       height;
	  string                        from;
	  void flushKV2DB(void);
};

} } // namespace eosio::chain

//FC_REFLECT(eosio::chain::apply_context::apply_results, (applied_actions)(deferred_transaction_requests)(deferred_transactions_count))
