PROJECT(wasm_serialize)

cmake_minimum_required(VERSION 3.5)

add_library( wasm_serialize
             token.cpp
	     token.hpp
             system.cpp
	     datastream.hpp            
             )

target_include_directories( wasm_serialize
			   PUBLIC "external/magic_get/include"
                           PUBLIC "./"
                           PUBLIC "../../external/magic_get/include"
                           PUBLIC "../adapter/include" 
			   PUBLIC "../fc/include/"
			   )
                           

target_link_libraries(wasm_serialize
		${Boost_LIBRARIES} )

install( TARGETS wasm_serialize
   RUNTIME DESTINATION ${CMAKE_INSTALL_FULL_BINDIR}
   LIBRARY DESTINATION ${CMAKE_INSTALL_FULL_LIBDIR}
   ARCHIVE DESTINATION ${CMAKE_INSTALL_FULL_LIBDIR}
)

