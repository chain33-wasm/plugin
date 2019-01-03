#pragma once

#include <eosio/chain/wasm_interface.hpp>
#include <eosio/chain/webassembly/wavm.hpp>
#include <eosio/chain/webassembly/binaryen.hpp>
#include <eosio/chain/webassembly/runtime_interface.hpp>
#include <eosio/chain/wasm_eosio_injection.hpp>
#include <eosio/chain/exceptions.hpp>
#include <fc/scoped_exit.hpp>

#include "IR/Module.h"
#include "Runtime/Intrinsics.h"
#include "Platform/Platform.h"
#include "WAST/WAST.h"
#include "IR/Validate.h"

using namespace fc;
using namespace eosio::chain::webassembly;
using namespace IR;
using namespace Runtime;

namespace eosio { namespace chain {

   struct wasm_interface_impl {
      wasm_interface_impl(wasm_interface::vm_type vm) {
         if(vm == wasm_interface::vm_type::wavm)
            runtime_interface = std::make_unique<webassembly::wavm::wavm_runtime>();
         else if(vm == wasm_interface::vm_type::binaryen)
            runtime_interface = std::make_unique<webassembly::binaryen::binaryen_runtime>();
         else
            EOS_THROW(wasm_exception, "wasm_interface_impl fall through");
      }

      std::vector<uint8_t> parse_initial_memory(const Module& module) {
	  	 //std::cout<<"\n\n*********parse_initial_memory\n";
         std::vector<uint8_t> mem_image;

         for(const DataSegment& data_segment : module.dataSegments) {
            EOS_ASSERT(data_segment.baseOffset.type == InitializerExpression::Type::i32_const, wasm_exception, "");
            EOS_ASSERT(module.memories.defs.size(), wasm_exception, "");
            const U32 base_offset = data_segment.baseOffset.i32;
            const Uptr memory_size = (module.memories.defs[0].type.size.min << IR::numBytesPerPageLog2);
            if(base_offset >= memory_size || base_offset + data_segment.data.size() > memory_size)
               FC_THROW_EXCEPTION(wasm_execution_error, "WASM data segment outside of valid memory range");
            if(base_offset + data_segment.data.size() > mem_image.size())
               mem_image.resize(base_offset + data_segment.data.size(), 0x00);
			
			////////////debug code///////
			char debugInfo[1024];
			int len = sprintf(debugInfo, "parse_initial_memory, memcpy base_offset:0x%x, size:%lu\n",
				base_offset, data_segment.data.size());
			//std::cout<<debugInfo<<"\n";
			len = 0;
			for (int j =0; j < data_segment.data.size() && len < 1024; j++) {
				len += sprintf(debugInfo + len, "%c", data_segment.data[j]);
			}
			//std::cout<<"data_segment.data: "<<debugInfo<<"\n";
            memcpy(mem_image.data() + base_offset, data_segment.data.data(), data_segment.data.size());
         }

		 //std::cout<<"\n************\nFinish parse_initial_memory\n";

         return mem_image;
      }

      std::unique_ptr<wasm_instantiated_module_interface>& get_instantiated_module( const digest_type& code_id,
                                                                                    const char *pcode,
	                                                                                int code_size)
      {
         fc::writewasmRunLog("Begin to do get_instantiated_module\n");
         auto it = instantiation_cache.find(code_id);
         if(it == instantiation_cache.end()) {
            IR::Module module;
            try {				
	           fc::writewasmRunLog("Begin to do WASM::serialize\n");
               Serialization::MemoryInputStream stream((const U8*)pcode, code_size);
               WASM::serialize(stream, module);
			   fc::writewasmRunLog("Finish to do WASM::serialize\n");
               module.userSections.clear();
            } catch(const Serialization::FatalSerializationException& e) {
               EOS_ASSERT(false, wasm_serialization_error, e.message.c_str());
            } catch(const IR::ValidationException& e) {
               EOS_ASSERT(false, wasm_serialization_error, e.message.c_str());
            }

			fc::writewasmRunLog("get_instantiated_module flag 1.\n");

            wasm_injections::wasm_binary_injection injector(module);
            injector.inject();

			fc::writewasmRunLog("get_instantiated_module flag 2.\n");

            std::vector<U8> bytes;
            try {
               Serialization::ArrayOutputStream outstream;
               WASM::serialize(outstream, module);
               bytes = outstream.getBytes();
            } catch(const Serialization::FatalSerializationException& e) {
               EOS_ASSERT(false, wasm_serialization_error, e.message.c_str());
            } catch(const IR::ValidationException& e) {
               EOS_ASSERT(false, wasm_serialization_error, e.message.c_str());
            }
			fc::writewasmRunLog("get_instantiated_module flag 3.\n");
            it = instantiation_cache.emplace(code_id, runtime_interface->instantiate_module((const char*)bytes.data(), bytes.size(), parse_initial_memory(module))).first;
         }
		 fc::writewasmRunLog("end to do get_instantiated_module\n");
         return it->second;
      }

      std::unique_ptr<wasm_runtime_interface> runtime_interface;
      map<digest_type, std::unique_ptr<wasm_instantiated_module_interface>> instantiation_cache;
   };

//Modify here to control intrinsic
#define _REGISTER_INTRINSIC_EXPLICIT(CLS, MOD, METHOD, WASM_SIG, NAME, SIG)\
   _REGISTER_WAVM_INTRINSIC(CLS, MOD, METHOD, WASM_SIG, NAME, SIG)\
   _REGISTER_BINARYEN_INTRINSIC(CLS, MOD, METHOD, WASM_SIG, NAME, SIG)

#define _REGISTER_INTRINSIC4(CLS, MOD, METHOD, WASM_SIG, NAME, SIG)\
   _REGISTER_INTRINSIC_EXPLICIT(CLS, MOD, METHOD, WASM_SIG, NAME, SIG )

#define _REGISTER_INTRINSIC3(CLS, MOD, METHOD, WASM_SIG, NAME)\
   _REGISTER_INTRINSIC_EXPLICIT(CLS, MOD, METHOD, WASM_SIG, NAME, decltype(&CLS::METHOD) )

#define _REGISTER_INTRINSIC2(CLS, MOD, METHOD, WASM_SIG)\
   _REGISTER_INTRINSIC_EXPLICIT(CLS, MOD, METHOD, WASM_SIG, BOOST_PP_STRINGIZE(METHOD), decltype(&CLS::METHOD) )

#define _REGISTER_INTRINSIC1(CLS, MOD, METHOD)\
   static_assert(false, "Cannot register " BOOST_PP_STRINGIZE(CLS) ":" BOOST_PP_STRINGIZE(METHOD) " without a signature");

#define _REGISTER_INTRINSIC0(CLS, MOD, METHOD)\
   static_assert(false, "Cannot register " BOOST_PP_STRINGIZE(CLS) ":<unknown> without a method name and signature");

#define _UNWRAP_SEQ(...) __VA_ARGS__

#define _EXPAND_ARGS(CLS, MOD, INFO)\
   ( CLS, MOD, _UNWRAP_SEQ INFO )

#define _REGISTER_INTRINSIC(R, CLS, INFO)\
   BOOST_PP_CAT(BOOST_PP_OVERLOAD(_REGISTER_INTRINSIC, _UNWRAP_SEQ INFO) _EXPAND_ARGS(CLS, "env", INFO), BOOST_PP_EMPTY())

#define REGISTER_INTRINSICS(CLS, MEMBERS)\
   BOOST_PP_SEQ_FOR_EACH(_REGISTER_INTRINSIC, CLS, _WRAPPED_SEQ(MEMBERS))

#define _REGISTER_INJECTED_INTRINSIC(R, CLS, INFO)\
   BOOST_PP_CAT(BOOST_PP_OVERLOAD(_REGISTER_INTRINSIC, _UNWRAP_SEQ INFO) _EXPAND_ARGS(CLS, EOSIO_INJECTED_MODULE_NAME, INFO), BOOST_PP_EMPTY())

#define REGISTER_INJECTED_INTRINSICS(CLS, MEMBERS)\
   BOOST_PP_SEQ_FOR_EACH(_REGISTER_INJECTED_INTRINSIC, CLS, _WRAPPED_SEQ(MEMBERS))

} } // eosio::chain
