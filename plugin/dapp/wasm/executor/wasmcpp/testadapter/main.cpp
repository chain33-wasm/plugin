#include <stdio.h>
#include <stdlib.h>
#include <eosio/chain/wasm_interface_adapter.h>
#include <string>
#include <iostream>
#include <sys/types.h>
#include <sys/stat.h>
#include <eosio/chain/_cgo_export.h>
#include <fc/io/json.hpp>
#include <eosio/chain/abi_def.hpp>

#include "../serialize/token.hpp"
#include "../serialize/serialize_api.h"



#if 1
int StateDBGetStateCallback(char* p0, char* p1, int p2, char*p3, int p4) {
    return 1;
}

void StateDBSetStateCallback(char* p0, char* p1, int p2, char*p3, int p4) {

}

int StateDBGetValueSizeCallback(char* p0, char* p1, int p2) {
    return 0;
}

void Output2UserCallback(char* p0, char* p1, int p2) {

}

int ExecFrozen(char *addr, long long int amount) {
	return 0;
}
int ExecActive(char *addr, long long int amount) {
	return 0;
}
int ExecTransfer(char *from, char *to, long long int amount) {
	return 0;
}
int ExecTransferFrozen(char *from, char *to, long long int amount) {
	return 0;
}

int GetRandom(char* randomDataOutput , int maxLen) {
    return 0;
}


#endif

using namespace std;

char *read_from_file(const char *filename, int *codeLen)
{
   FILE * pFile = NULL;
   char *buffer = nullptr;
   pFile = fopen (filename , "rb");
   if (pFile == NULL) {
      perror ("Error opening file");
   } else {

     struct stat info;  
	 stat(filename, &info);  
	 int size = info.st_size; 
	 
	 printf("size of file:%s is %d.\n", filename, size);
	 buffer = (char*) malloc (sizeof(char)*size);

	 int result = fread (buffer,1,size ,pFile);
     if (result != size) {
	 	fputs ("Reading error", stderr); 
		exit (3);
	}
    *codeLen = result;
    fclose (pFile);
   }
   return buffer;
}

void printHelp(void) {
	printf("pls run the programme due to your purpose...\n");
	printf("validation        : ./test_adapter validate wasmfile\n");
	printf("unserializeViaAbi : ./test_adapter unserializeViaAbi abifile \n");
}

void validate(char *wasmFilePath) {
    int codelen = 0;
    char *code = NULL;
    code = read_from_file(wasmFilePath, &codelen);
    //begin to print the upper part of the wasm file
#if 0
    int readlen = 20;
    printf("The first %d char is:", readlen);
    for (int i = 0; i < readlen; i++) {
        printf(" %02x", code[i]);
    }
    printf("\n");
#endif

    int result = wasm_validate_contract(code, codelen);
    printf("Run wasm contract validation result is:%d\n", result);
	if (NULL != code) {
        free(code);
	}    
}

void unserializeViaAbi(char *abiFilePath) {
    int abilen = 0;
    char *abi = NULL;
	
    abi = read_from_file(abiFilePath, &abilen);

	char data[512] = {0};
    account accinfoOrigin;
	accinfoOrigin.amount = 1009;
	accinfoOrigin.symbol = string("YCC");
	int value_size = serializeAccount(data, 512, accinfoOrigin);
#if 1
    printf("Result of serializeAccount length:%d\n", value_size);
    for (int i = 0; i < value_size; i++) {
        printf(" %02x", (char)data[i]);
		//printf(" %d", data[i]);
    }
    printf("\n");
#endif

	char *pJsonResult = NULL;
	int result = convertData2Json((const char *)abi, 
	                      data, value_size, 
	                      "accounts", &pJsonResult);
	if (result != 0) {
		printf("Fail to do operation unserializeViaAbi\n");
		return;
	}
	printf("^-^ ^-^ ^-^ Succeed.\n");
	printf("The result with json is as below:\n %s \n", pJsonResult);
	free(pJsonResult);	
}


void callContract(const char *wasmFilePath) {

}


int main( int argc, char** argv ) {
	if (argc == 3) {
		std::string actiontype(argv[1]);
		if (0 == actiontype.compare("validate")) {
			validate(argv[2]);
		} else if (0 == actiontype.compare("unserializeViaAbi")) {
		    unserializeViaAbi(argv[2]);
		}
	} else {
	    printHelp();
	}
	return 0;
}



