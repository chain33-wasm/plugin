# CMAKE generated file: DO NOT EDIT!
# Generated by "Unix Makefiles" Generator, CMake Version 3.13

# Delete rule output on recipe failure.
.DELETE_ON_ERROR:


#=============================================================================
# Special targets provided by cmake.

# Disable implicit rules so canonical targets will work.
.SUFFIXES:


# Remove some rules from gmake that .SUFFIXES does not remove.
SUFFIXES =

.SUFFIXES: .hpux_make_needs_suffix_list


# Suppress display of executed commands.
$(VERBOSE).SILENT:


# A target that is always out of date.
cmake_force:

.PHONY : cmake_force

#=============================================================================
# Set environment variables for the build.

# The shell in which to execute make rules.
SHELL = /bin/sh

# The CMake executable.
CMAKE_COMMAND = /usr/local/bin/cmake

# The command to remove a file.
RM = /usr/local/bin/cmake -E remove -f

# Escaping for special characters.
EQUALS = =

# The top-level source directory on which CMake was run.
CMAKE_SOURCE_DIR = /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp

# The top-level build directory on which CMake was run.
CMAKE_BINARY_DIR = /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp

# Utility rule file for Inline.

# Include the progress variables for this target.
include wasm-jit/Include/Inline/CMakeFiles/Inline.dir/progress.make

Inline: wasm-jit/Include/Inline/CMakeFiles/Inline.dir/build.make

.PHONY : Inline

# Rule to build all files generated by this target.
wasm-jit/Include/Inline/CMakeFiles/Inline.dir/build: Inline

.PHONY : wasm-jit/Include/Inline/CMakeFiles/Inline.dir/build

wasm-jit/Include/Inline/CMakeFiles/Inline.dir/clean:
	cd /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp/wasm-jit/Include/Inline && $(CMAKE_COMMAND) -P CMakeFiles/Inline.dir/cmake_clean.cmake
.PHONY : wasm-jit/Include/Inline/CMakeFiles/Inline.dir/clean

wasm-jit/Include/Inline/CMakeFiles/Inline.dir/depend:
	cd /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp && $(CMAKE_COMMAND) -E cmake_depends "Unix Makefiles" /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp/wasm-jit/Include/Inline /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp/wasm-jit/Include/Inline /home/yann/go/src/github.com/33cn/plugin/plugin/dapp/wasm/executor/wasmcpp/wasm-jit/Include/Inline/CMakeFiles/Inline.dir/DependInfo.cmake --color=$(COLOR)
.PHONY : wasm-jit/Include/Inline/CMakeFiles/Inline.dir/depend
