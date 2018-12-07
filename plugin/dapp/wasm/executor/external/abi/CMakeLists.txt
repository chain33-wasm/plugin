# copyright defined in abieos/LICENSE.txt

cmake_minimum_required (VERSION 3.11)
project(abiconv VERSION 0.1 LANGUAGES CXX)

FIND_PACKAGE(Boost 1.58 REQUIRED COMPONENTS date_time)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_library(abiconv STATIC src/abieos.cpp)
target_include_directories(abiconv PUBLIC external/rapidjson/include PRIVATE ${Boost_INCLUDE_DIR})
target_link_libraries(abiconv Boost::date_time)


if (CMAKE_CXX_COMPILER_ID MATCHES "Clang")
    target_compile_options(abiconv PRIVATE -Wall -Wextra -Wno-unused-parameter)
endif()
