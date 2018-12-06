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







