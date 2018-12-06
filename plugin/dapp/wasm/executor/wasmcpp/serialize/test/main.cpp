#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <iostream>
//#include "../token.h"
#include "../token.hpp"
#include "../serialize_api.h"

using namespace std;

int main( int argc, char** argv ) {
    char data[512] = {0};
    account accinfoOrigin;
	accinfoOrigin.amount = 1009;
	accinfoOrigin.symbol = string("YCC");
	int size = serializeAccount(data, 512, accinfoOrigin);  
	
	
	account4Go accinfo;
	unserializeAccount(data, size, &accinfo);

	printf("sizeof =%lu, pack_size=%d\n", 
		   sizeof(accinfoOrigin), size);

	printf("eosio::unserializeAccount is called with result:%ld, %s\n", 
		accinfo.amount, accinfo.symbol);
}




