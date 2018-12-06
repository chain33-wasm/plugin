
file(GLOB HEADERS "include/eosio/chain/*.hpp")
#add_definitions(-D_GLIBCXX_USE_CXX11_ABI=0)
## SORT .cpp by most likely to change / break compile
add_executable( test_serialize
             main.cpp
             )
target_link_libraries(test_serialize 
         PRIVATE wasm_serialize fc  
                     )

target_include_directories(test_serialize
                           PUBLIC "../"
                           PUBLIC "../../external/magic_get/include")

