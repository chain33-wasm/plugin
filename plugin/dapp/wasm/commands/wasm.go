package commands

import (
	"encoding/json"
	"fmt"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func WasmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "WASM contracts operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		//
		wasmCheckContractNameCmd(),
		wasmCreateContractCmd(),
		wasmGenAbiCmd(),
		wasmCallContractCmd(),
		wasmQueryContractCmd(),
		wasmEstimateContractCmd(),
		wasmDebugCmd(),
	)

	return cmd
}

// 创建wasm合约
func wasmCreateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new WASM contract",
		Run:   wasmCreateContract,
	}
	wasmAddCreateContractFlags(cmd)
	return cmd
}

func wasmAddCreateContractFlags(cmd *cobra.Command) {
	wasmAddCommonFlags(cmd)

	cmd.Flags().StringP("contract", "x", "", "contract name same with the code and abi file")
	cmd.MarkFlagRequired("contract")

	cmd.Flags().StringP("path", "d", "", "path where stores wasm code and abi")
	cmd.MarkFlagRequired("path")
}

func wasmCreateContract(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("contract")
	path, _ := cmd.Flags().GetString("path")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	//paraName, _ := cmd.Flags().GetString("paraName")

	feeInt64 := uint64(fee*1e4) * 1e4

	codePath := path + "/" + contractName + ".wasm"
	abiPath := path + "/" + contractName + ".abi"
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read code error ", err)
		return
	}

	abi, err := ioutil.ReadFile(abiPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read abi error ", err)
		return
	}

	value := wasmtypes.CreateWasmContractPara{
		Code: code,
		Abi:  string(abi),
		Name: contractName,
		Note: note,
		Fee:  int64(feeInt64),
	}

	paramJson, errInfo := json.Marshal(value)
	if errInfo != nil {
		fmt.Fprintln(os.Stderr, "json.Marshal error ", errInfo)
		return

	}

	params := &rpctypes.CreateTxIn{
		Execer:     wasmtypes.WasmX,
		ActionName: wasmtypes.CreateWasmContractStr,
		Payload:    paramJson,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func wasmGenAbiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abi data",
		Short: "Generate abi data",
		Run:   wasmGenAbiData,
	}
	wasmAddGenAbiDataFlags(cmd)
	return cmd
}

func wasmAddGenAbiDataFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "external contract name, like user.external.xxxxx")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("action", "a", "", "action name")
	cmd.MarkFlagRequired("action")

	cmd.Flags().StringP("data", "d", "", "action data in json string")
	cmd.MarkFlagRequired("data")
}

func wasmGenAbiData(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	contractAddr, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	actionData, _ := cmd.Flags().GetString("data")

	req := types.ReqAddr{Addr: contractAddr}
	var res wasmtypes.WasmGetAbiResp
	query := sendQuery4wasm(rpcLaddr, "WasmGetAbi", &req, &res)
	if query {
		abidata := genAbiData(string(res.Abi), contractAddr, actionName, actionData)
		fmt.Println(string("The converted abi data is:") + common.ToHex(abidata))
	} else {
		fmt.Fprintln(os.Stderr, "get abi data error")
	}
}

//运行wasm合约的查询请求
func wasmQueryContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query the WASM contract for specified table ",
		Run:   wasmQueryContract,
	}
	wasmAddQueryContractFlags(cmd)
	return cmd
}

func wasmQueryContract(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("exec")
	tableName, _ := cmd.Flags().GetString("table")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	queryReq := wasmtypes.WasmQueryContractTableReq{
		ContractName: contractName,
		TableName:    []string{tableName},
	}

	var WasmQueryResponse wasmtypes.WasmQueryResponse
	query := sendQuery4wasm(rpcLaddr, wasmtypes.QueryFromContract, &queryReq, &WasmQueryResponse)
	if query {
		for _, WasmOutItem := range WasmQueryResponse.QueryResultItems {
			fmt.Println(WasmOutItem.ItemType)
			fmt.Println(WasmOutItem.ResultJSON)
		}
	} else {
		fmt.Fprintln(os.Stderr, "get wasm query error")
		return
	}
}

func wasmAddQueryContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "wasm contract name")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("table", "n", "", "one of wasm contract's table name")
	cmd.MarkFlagRequired("table")
}

// 调用WASM合约
func wasmCallContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "Call the WASM contract",
		Run:   wasmCallContract,
	}
	wasmAddCallContractFlags(cmd)
	return cmd
}

func wasmCallContract(cmd *cobra.Command, args []string) {
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	contractName, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	abiPara, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	feeInt64 := uint64(fee*1e4) * 1e4

	req := wasmtypes.ConvertJson2AbiReq{
		ContractName:  contractName,
		ActionName:    actionName,
		AbiDataInJson: abiPara,
	}
	var json2AbiResponse wasmtypes.Json2AbiResponse
	var actionData []byte
	query := sendQuery4wasm(rpcLaddr, wasmtypes.ConvertJson2Abi, &req, &json2AbiResponse)
	if query {
		actionData = json2AbiResponse.AbiData
	} else {
		fmt.Fprintln(os.Stderr, "get abi data error")
		return
	}

	value := wasmtypes.CallWasmContractPara{
			Name:       contractName,
			Note:       note,
			ActionName: actionName,
			ActionData: actionData,
			Fee:        int64(feeInt64)}

	paramJson, errInfo := json.Marshal(value)
	if errInfo != nil {
		fmt.Fprintln(os.Stderr, "json.Marshal error ", errInfo)
		return

	}

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(wasmtypes.WasmX),
		ActionName: wasmtypes.CallWasmContractStr,
		Payload:    paramJson,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func wasmAddCallContractFlags(cmd *cobra.Command) {
	wasmAddCommonFlags(cmd)
	cmd.Flags().StringP("exec", "e", "", "wasm contract name,like user.wasm.xxx")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("action", "x", "", "external contract action name")
	cmd.MarkFlagRequired("action")

	cmd.Flags().StringP("para", "r", "", "external contract execution parameter in json string")
	cmd.MarkFlagRequired("para")
}

func wasmAddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func wasmEstimateContract(cmd *cobra.Command, args []string) {
	code, _ := cmd.Flags().GetString("input")
	name, _ := cmd.Flags().GetString("exec")
	caller, _ := cmd.Flags().GetString("caller")
	amount, _ := cmd.Flags().GetFloat64("amount")

	toAddr := address.ExecAddress("external")
	if len(name) > 0 {
		toAddr = address.ExecAddress(name)
	}

	amountInt64 := uint64(amount*1e4) * 1e4
	bCode, err := common.FromHex(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse external code error", err)
		return
	}

	var estGasReq = wasmtypes.EstimateWASMGasReq{To: toAddr, Code: bCode, Caller: caller, Amount: amountInt64}
	var estGasResp wasmtypes.EstimateWASMGasResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4wasm(rpcLaddr, "EstimateGas", &estGasReq, &estGasResp)

	if query {
		fmt.Fprintf(os.Stdout, "gas cost estimate %v\n", estGasResp.Gas)
	} else {
		fmt.Fprintln(os.Stderr, "gas cost estimate error")
	}
}

func addEstimateFlags4wasm(cmd *cobra.Command) {
	cmd.Flags().StringP("input", "i", "", "input contract binary code")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("exec", "e", "", "external contract name (like user.external.xxxxx)")

	cmd.Flags().StringP("caller", "c", "", "the caller address")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")
}

func addEstimateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("input", "i", "", "input contract binary code")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("exec", "e", "", "evm contract name (like user.evm.xxxxx)")

	cmd.Flags().StringP("caller", "c", "", "the caller address")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")
}

// 估算合约消耗
func wasmEstimateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate the gas cost of calling or creating a contract",
		Run:   wasmEstimateContract,
	}
	addEstimateFlags(cmd)
	return cmd
}

// 检查地址是否为WASM合约
func wasmCheckContractNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if wasm contract used has been used already",
		Run:   wasmCheckContractAddr,
	}
	wasmAddCheckContractAddrFlags(cmd)
	return cmd
}

func wasmAddCheckContractAddrFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "wasm contract name, like user.wasm.xxxxx(a-z0-9, within length [4-16])")
	cmd.MarkFlagRequired("exec")
}

func wasmCheckContractAddr(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("exec")

	var checkAddrReq = wasmtypes.CheckWASMContractNameReq{WasmContractName: name}
	var checkAddrResp wasmtypes.CheckWASMAddrResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4wasm(rpcLaddr, wasmtypes.CheckNameExistsFunc, &checkAddrReq, &checkAddrResp)

	if query {
		proto.MarshalText(os.Stdout, &checkAddrResp)
	} else {
		fmt.Fprintln(os.Stderr, "error")
	}
}

// 查询或设置WASM调试开关
func wasmDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Query or set external debug status",
	}
	cmd.AddCommand(
		wasmDebugQueryCmd(),
		wasmDebugSetCmd(),
		wasmDebugClearCmd())

	return cmd
}

func wasmDebugQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Query external debug status",
		Run:   wasmDebugQuery,
	}
}
func wasmDebugSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set external debug to ON",
		Run:   wasmDebugSet,
	}
}
func wasmDebugClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Set external debug to OFF",
		Run:   wasmDebugClear,
	}
}

func wasmDebugQuery(cmd *cobra.Command, args []string) {
	wasmDebugRpc(cmd, 0)
}

func wasmDebugSet(cmd *cobra.Command, args []string) {
	wasmDebugRpc(cmd, 1)
}

func wasmDebugClear(cmd *cobra.Command, args []string) {
	wasmDebugRpc(cmd, -1)
}
func wasmDebugRpc(cmd *cobra.Command, flag int32) {
	var debugReq = wasmtypes.WasmDebugReq{Optype: flag}
	var debugResp wasmtypes.WasmDebugResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4wasm(rpcLaddr, "WasmDebug", &debugReq, &debugResp)

	if query {
		proto.MarshalText(os.Stdout, &debugResp)
	} else {
		fmt.Fprintln(os.Stderr, "error")
	}
}

func sendQuery4wasm(rpcAddr, funcName string, request types.Message, result proto.Message) bool {
	params := rpctypes.Query4Jrpc{
		Execer:   wasmtypes.WasmX,
		FuncName: funcName,
		Payload:  types.MustPBToJSON(request),
	}

	jsonrpc, err := jsonclient.NewJSONClient(rpcAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = jsonrpc.Call("Chain33.Query", params, result)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}
