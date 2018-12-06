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



//#ifdef __cplusplus
//}
//#endif


//#endif


