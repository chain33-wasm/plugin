
export BOOST_ROOT="${HOME}/opt/boost"
#CMAKE_BUILD_TYPE=Debug
CMAKE_BUILD_TYPE=Release
CXX_COMPILER=clang++-4.0
C_COMPILER=clang-4.0
CORE_SYMBOL_NAME="SYS"
OPENSSL_ROOT_DIR=/usr/include/openssl
NABLE_COVERAGE_TESTING=false
DOXYGEN=false
SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "source dir is:"${SOURCE_DIR}

BUILD_DIR="${PWD}/build"
if [ ! -d "${BUILD_DIR}" ]; then
      if ! mkdir -p "${BUILD_DIR}"
      then
         printf "Unable to create build directory %s.\\n Exiting now.\\n" "${BUILD_DIR}"
         exit 1;
      fi  
   fi  

   if ! cd "${BUILD_DIR}"
   then
      printf "Unable to enter build directory %s.\\n Exiting now.\\n" "${BUILD_DIR}"
      exit 1;
fi  
echo "sleep 3 seconds and clean build"
#rm * -fr

COUNTER=0
while [[ ${COUNTER} -lt 3 ]]
do
    ((COUNTER++))
    echo "."
    sleep 1
done

if [ -z "$CMAKE" ]; then
   CMAKE=$( command -v cmake )
fi

if ! "${CMAKE}" -DCMAKE_BUILD_TYPE="${CMAKE_BUILD_TYPE}" -DCMAKE_CXX_COMPILER="${CXX_COMPILER}" \
      -DCMAKE_C_COMPILER="${C_COMPILER}" -DWASM_ROOT="${WASM_ROOT}" -DCORE_SYMBOL_NAME="${CORE_SYMBOL_NAME}" \
      -DOPENSSL_ROOT_DIR="${OPENSSL_ROOT_DIR}" -DBUILD_MONGO_DB_PLUGIN=true \
      -DENABLE_COVERAGE_TESTING="${ENABLE_COVERAGE_TESTING}" -DBUILD_DOXYGEN="${DOXYGEN}" \
      -DCMAKE_INSTALL_PREFIX="/usr/local/eosio" "${SOURCE_DIR}"
then
      printf "\\n\\t>>>>>>>>>>>>>>>>>>>> CMAKE building EOSIO has exited with the above error.\\n\\n"
      exit -1
fi
