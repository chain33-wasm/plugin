#include <algorithm>
#include <eosio/chain/apply_context.hpp>
#include <eosio/chain/exceptions.hpp>
#include <eosio/chain/wasm_interface.hpp>
#include <eosio/chain/account_object.hpp>
#include <eosio/chain/callback2Go.h>

using boost::container::flat_set;

namespace eosio { namespace chain {
uint64_t SstoreSetGas      = 20000; 
uint64_t SstoreResetGas    = 5000;  
uint64_t SstoreClearGas    = 5000;
uint64_t SstoreLoadGas     = 50;

/*
static inline void print_debug(account_name receiver, const action_trace& ar) {
   if (!ar.console.empty()) {
      auto prefix = fc::format_string(
                                      "\n[(${a},${n})->${r}]",
                                      fc::mutable_variant_object()
                                      ("a", ar.act.account)
                                      ("n", ar.act.name)
                                      ("r", receiver));
      dlog(prefix + ": CONSOLE OUTPUT BEGIN =====================\n"
           + ar.console
           + prefix + ": CONSOLE OUTPUT END   =====================" );
   }
}
*/

apply_context::~apply_context() {
    std::string temp("Apply context console output is:\n");
    temp += _pending_console_output.str(); 
	fc::writewasmRunLog(temp.c_str());
	char buf[512] = {0};
	sprintf(buf, "There are total %d element need to write back", out2user.size());
	fc::writewasmRunLog(buf);
	for (vector<Out2UserElement>::iterator it = out2user.begin();
	     it != out2user.end();
		 it++) {
		 fc::writewasmRunLog(it->type.c_str());
		 Output2UserViaBlockchain((char *)it->type.c_str(), it->data.data(), it->data.size());
	}
	
}


void apply_context::reset_console() {
   _pending_console_output = std::ostringstream();
   _pending_console_output.setf( std::ios::scientific, std::ios::floatfield );
}

gas_check_res apply_context::check_and_spend_gas(int gascost) {
	gasAvailable -= gascost;
	if (gasAvailable >= 0) {
		return GAS_CHECK_SUCCEES;
	}
	EOS_ASSERT(false, gas_usage4normal_exceeded , "gasAvailable:${a}, gascost:${b}",
	          ("a", gasAvailable)("b", gascost));
	return GAS_CHECK_FAIL;
}

void apply_context::output2user(const char *type, const char *data, int len) {
	Out2UserElement out2UserElement;
	out2UserElement.type = string(type);
	out2UserElement.data.resize(len, 0);
	memcpy(out2UserElement.data.data(), data, len);
	out2user.emplace_back(out2UserElement);
	//////////////////debug code////////////////////////
	char temp[512] = {0};
	sprintf(temp, "structure:%s is added with size:%d", type, out2user.size());
	fc::writewasmRunLog(temp);
#if 0
	if (0 == out2user.size()) {
		out2user.resize(len, 0);
		memcpy(out2user.data(), data, len);
	} else {
	    int total = len + out2user.size();
		//To keep completeness of data,just save data slice not exceed
		if (total > MAX_OUT2USER_SIZE) {
			fc::writewasmRunLog("Info from output2user exceeds 1M\n");
		}

        int begin = out2user.size();
		key4DB.resize(total, 0x00);
		memcpy(out2user.data() + begin, data, len);
	}
#endif
}


void apply_context::setKey(const char *data, int len) {
	key4DB.resize(len, 0x00);
	memset(key4DB.data(), 0x00, len);
	memcpy(key4DB.data(), data, len);
}
void apply_context::setValue(const char *data, int len) {
	value4DB.resize(len, 0x00);
	memset(value4DB.data(), 0x00, len);
	memcpy(value4DB.data(), data, len);
	flushKV2DB();
}

int apply_context::getValueSize(const char *data, int len) {
	key4DB.resize(len, 0x00);
	memset(key4DB.data(), 0x00, len);
	memcpy(key4DB.data(), data, len);
	return StateDBGetValueSize(&contractAddr[0], key4DB.data(), len);
}

int apply_context::getValue(char *data, int len) {
	return StateDBGetState(&contractAddr[0], (char *)key4DB.data(), key4DB.size(), data, len);
}

void apply_context::flushKV2DB(void) {
    char *ppvalue = NULL;
	int vallenNow = 0;
	int vallen = value4DB.size();
	
    vallenNow = StateDBGetValueSize(&contractAddr[0], (char *)key4DB.data(), key4DB.size());
		
	int64_t gas_need = 0;
	//set
	if ((vallenNow == 0) && (vallen != 0)) {
		gas_need = SstoreSetGas * (vallen + 31) / 32;
	} else if ((vallenNow != 0) && (vallen != 0)) { //reset
	    int minval = 0;
		minval = vallen < vallenNow ? vallen : vallenNow;
	    gas_need = SstoreResetGas * (minval + 31) / 32;
		if (vallen > minval) {
			gas_need += SstoreSetGas * (vallen - minval + 31) / 32;
		}
	} else if (vallen == 0) { //clear
		gas_need = SstoreClearGas;
	}
	check_and_spend_gas(gas_need);			
	
	StateDBSetState(&contractAddr[0], (char *)key4DB.data(), key4DB.size(), value4DB.data(), value4DB.size());
	free(ppvalue);
}

int64_t apply_context::getBlockTime()const {
	return currentBlockTime;
}

int apply_context::get_from(char *fromAddr, size_t from_size) {
	int min = std::min(from_size, from.size());
	memcpy(fromAddr, from.c_str(), from.size());

	return min;
}

////////////functions for token operations//////////////////////
int apply_context::execFrozenCoin(char* addr, long long int p1) {
    return execFrozen(addr, p1);
}

int apply_context::execActiveCoin(char* addr, long long int p1) {
	return execActive(addr, p1);
}

int apply_context::execTransferCoin(char* from, char* to, long long int p2) {
	return execTransfer(from, to, p2);
}

int apply_context::execTransferFrozenCoin(char* from, char* to, long long int p2) {
    return execTransferFrozen(from, to, p2);
}

} } /// eosio::chain
