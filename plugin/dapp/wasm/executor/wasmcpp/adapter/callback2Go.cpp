//import go callback functions
#include <eosio/chain/_cgo_export.h>
#include <eosio/chain/callback2Go.h>

int StateDBGetState(char* p0, char* p1, int p2, char*p3, int p4) {
    return StateDBGetStateCallback(p0, p1, p2, p3, p4);

}
void StateDBSetState(char* p0, char* p1, int p2, char* p3, int p4) {
	StateDBSetStateCallback(p0, p1, p2, p3, p4);
}

int StateDBGetValueSize(char* p0, char* p1, int p2) {
	return StateDBGetValueSizeCallback(p0, p1, p2);
}

void Output2UserViaBlockchain(char *p0, char* p1, int p2) {
	Output2UserCallback(p0, p1, p2);
}

int execFrozen (char* addr, long long int p1) {
	return ExecFrozen(addr, p1);
}


int execActive (char* addr, long long int p1) {
	return ExecActive(addr, p1);
}

int execTransfer (char* from, char* to, long long int p2) {
	return ExecTransfer(from, to, p2);

}

int execTransferFrozen (char* from, char* to, long long int p2) {
	return ExecTransferFrozen(from, to, p2);
}

int getRandom(char* randomDataOutput , int maxLen) {
    return GetRandom(randomDataOutput , maxLen);
}









