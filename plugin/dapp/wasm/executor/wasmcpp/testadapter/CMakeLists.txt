
file(GLOB HEADERS "include/eosio/chain/*.hpp")
#add_definitions(-D_GLIBCXX_USE_CXX11_ABI=0)
## SORT .cpp by most likely to change / break compile
add_executable( test_adapter
             main.cpp
             )
target_link_libraries(test_adapter 
         PRIVATE wasm_adapter wasm_serialize fc chainbase Logging IR WAST WASM Runtime wasm asmjs passes cfg ast emscripten-optimizer support softfloat builtins
                     )
