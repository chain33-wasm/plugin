#include <eosio/chain/wasm_interface.hpp>
#include <eosio/chain/wasm_interface_adapter.h>
#include <eosio/chain/abi_serializer.hpp>
#include <eosio/chain/action.hpp>
#include <eosio/chain/apply_context.hpp>
#include <fc/io/json.hpp>
#include <stdio.h>

//using namespace webassembly;
//using namespace webassembly::common;

//using vm_type = chain33::wasmcontract::wasm_interface::vm_type;
eosio::chain::wasm_interface::vm_type wasm_interface_vm_type_binaryen = eosio::chain::wasm_interface::vm_type::binaryen;
int VMTypeBinaryen = int(eosio::chain::wasm_interface::vm_type::binaryen);

//contract code must be validated before deployed
Result wasm_validate_contract(const char *pcode, int len) {
    std::string begin = fc::format_string("Begin to do wasm_validate_contract with len:${a}.\n",
		fc::mutable_variant_object()("a", len));
	
	fc::writewasmRunLog(begin.c_str());

	eosio::chain::bytes code;
	for(int i = 0; i < len; i++) {
		code.emplace_back(*pcode++);
	}

	try {
		eosio::chain::wasm_interface::validate(code);
		fc::writewasmRunLog("Succeed to do wasm_validate_contract\n");
		return Success;			
	} catch (const fc::exception& e) {
	    std::string excpInfo = fc::format_string("Failed to do wasm_validate_contract with "
			"friendly exception info :${a}\n",fc::mutable_variant_object()
			("a", e.to_detail_string()));
	    
		fc::writewasmRunLog(excpInfo.c_str());
	    return Validate_fail;
	}
}

//create apply context for contract execution
void create_apply_context(Apply_context_para *pApply_context) {
	(void)pApply_context;

#if 0
	std::cout<<"create_apply_context datalen:"<<pApply_context->datalen<<"\n";
    eosio::chain::bytes data(pApply_context->datalen);
	for(eosio::chain::bytes::iterator iter = data.begin(); iter != data.end(); iter++) {
        *iter = *pApply_context->pdata++;
	}
	eosio::chain::action act(eosio::chain::string_to_name(&pApply_context->contractName[0]), eosio::chain::string_to_name(&pApply_context->action_name[0]), data);
	if (NULL != pcontext) {
		delete pcontext;
		pcontext = NULL;
	}
	
	std::string contractAddrInStr = std::string(pApply_context->contractAddr);
	std::string from = std::string(pApply_context->from);
	pcontext = new eosio::chain::apply_context(pApply_context);	
#endif
}

//call contract with specified code and context
int callContract4go(int vm, const char *pcode, int code_size, Apply_context_para *pApply_context){
	std::unique_ptr<class eosio::chain::apply_context> pcontext(new eosio::chain::apply_context(pApply_context));
	try {
		fc::writewasmRunLog("Begin to do callContract4go\n");
		eosio::chain::digest_type code_id = eosio::chain::digest_type::hash(pcode, (uint32_t)code_size); 
		eosio::chain::wasm_interface *wasmIntInstance = new eosio::chain::wasm_interface(eosio::chain::wasm_interface::vm_type(vm));
		wasmIntInstance->apply(code_id, pcode, code_size, *pcontext);
		fc::writewasmRunLog("Successfully finished doing callContract4go\n");
		pApply_context->gasAvailable = pcontext->gasAvailable;

		return Success;
	} catch (const fc::exception& e){
	    //throw;
	    //FC_CAPTURE_AND_RETHROW((trace))
		std::string excpInfo = fc::format_string("Failed to do callContract4go due to "
			"friendly exception info :${a}\n",fc::mutable_variant_object()
			("a", e.to_detail_string()));
	    
		fc::writewasmRunLog(excpInfo.c_str());
		pApply_context->gasAvailable = pcontext->gasAvailable;
	    if (pcontext->gasAvailable < 0) {
			return OUT_GAS;
		}
	    return Fail_exception;
    }		
}

int convertData2Json(const char *abiStr,
	                        const char *pvalue, int value_size,
	                        const char *table,
	                        char **ppJsonResult) {
	std::string abiString(abiStr);
	eosio::chain::abi_def abi = fc::json::from_string(abiString).as<eosio::chain::abi_def>();

    /////////////////////////////////////////////////////////////////
    //TODO: Consider using std::vector<char> packabi_string as input parameter
    //to generate abi_def
#if 0
	std::vector<char> packabi_string       = fc::raw::pack(abi);
	eosio::chain::abi_serializer::to_abi(packabi_string, abi);
#endif
	////////////////////////////////////////////////////////////////
	const fc::microseconds abi_serializer_max_time(1000 * 10);
	eosio::chain::abi_serializer abis;
	abis.set_abi(abi, abi_serializer_max_time);

	eosio::chain::bytes data;
	for(int i = 0; i < value_size; i++) {
		data.emplace_back(*pvalue++);
	}

	eosio::chain::type_name table_name = abis.get_table_type(eosio::chain::string_to_name(table));
	fc::variant result = abis.binary_to_variant(table_name, data, abi_serializer_max_time);
	std::string result_in_json = fc::json::to_pretty_string(result);

	int size = result_in_json.size();
	*ppJsonResult = (char *)malloc(size + 1);
	if (nullptr == *ppJsonResult) {
		return -1;
	}
	//make sure the string end with '\0'
	((char *)(*ppJsonResult))[size] = 0;
	memcpy(*ppJsonResult, result_in_json.c_str(), result_in_json.size());

	return 0;
}


