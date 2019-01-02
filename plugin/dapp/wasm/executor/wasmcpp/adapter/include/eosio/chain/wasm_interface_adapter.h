#ifndef WASM_INTERFACE_ADAPTER_H
#define WASM_INTERFACE_ADAPTER_H

// __cplusplus gets defined when a C++ compiler processes the file
#ifdef __cplusplus
// extern "C" is needed so the C++ compiler exports the symbols without name
// manging.
extern "C" {
#endif

#define Success (0L)
#define Validate_fail (-1)
#define Fail_exception (1)
#define OUT_GAS       (2)
#define WRONG_PARA    (3)
#define Malloc_Fail   (4)



typedef int Result;	

typedef struct Apply_context_para {
	char *contractAddr;	
	char *contractName; /*alias*/
	char *action_name; 
	char *pdata;
	int datalen;
	char *from;
	char *loginfo;
	int64_t gasAvailable;
	int64_t blocktime;
	int64_t height;
} Apply_context_para;

    extern int VMTypeBinaryen;

    //contract code must be validated before deployed
	extern Result wasm_validate_contract(const char *pcode, int len);

    //create apply context for contract execution
	extern void create_apply_context(Apply_context_para *pApply_context);

    //call contract with specified code and context
	extern int callContract4go(int vm, 
				                          const char *pcode,
				                          int code_size,
				                          Apply_context_para *pApply_context);
    //unserialize data and converte it to json format
    extern int convertData2Json(const char *abi,
                                        const char *pvalue, int value_size,
                                        const char *table,
                                        char **ppJsonResult);

	


#ifdef __cplusplus
}
#endif


#endif
