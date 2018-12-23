//#ifndef WASM_CALLBACK2GO_H
//#define WASM_CALLBACK2GO_H

// __cplusplus gets defined when a C++ compiler processes the file
//#ifdef __cplusplus
// extern "C" is needed so the C++ compiler exports the symbols without name
// manging.
//extern "C" {
//#endif

#pragma once


int StateDBGetState(char* p0, char* p1, int p2, char*p3, int p4);
void StateDBSetState(char* p0, char* p1, int p2, char* p3, int p4);
int StateDBGetValueSize(char* p0, char* p1, int p2);
void Output2UserViaBlockchain(char *p0, char* p1, int p2);

int execFrozen(char* addr, long long int p1);
int execActive(char* addr, long long int p1);
int execTransfer(char* from, char* to, long long int p2);
int execTransferFrozen(char* from, char* to, long long int p2);

int getRandom(char* randomDataOutput , int maxLen);

//interface corresponding to local db operation
int getValueSizeFromLocal (char* p0, char* p1, int p2);
int getValueFromLocal (char* p0, char* p1, int p2, char* p3, int p4);
void setValue2Local (char* p0, char* p1, int p2, char* p3, int p4);

//#ifdef __cplusplus
//}
//#endif


//#endif


