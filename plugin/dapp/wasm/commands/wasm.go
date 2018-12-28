package commands

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"regexp"
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
		wasmCallContractCmd(),
		wasmQueryContractCmd(),
		wasmFuzzyQueryContractCmd(),
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

	nameReg, err := regexp.Compile(wasmtypes.NameRegExp)
	if !nameReg.MatchString(contractName) {
		fmt.Fprintln(os.Stderr, "Wrong wasm contract name format, which should be a-z and 0-9 ")
		return
	}

	if len(contractName) > 16 || len(contractName) < 4 {
		fmt.Fprintln(os.Stderr, "wasm contract name's length should be within range [4-16]")
		return
	}

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

	var createReq = wasmtypes.CreateContrantReq{
		Name: contractName,
		Note: note,
		Code: code,
		Abi:  string(abi),
		Fee:  int64(feeInt64),
	}
	var createResp = types.ReplyString{}
	query := sendQuery4wasm(rpcLaddr, wasmtypes.CreateWasmContractStr, &createReq, &createResp)
	if query {
		fmt.Println(createResp.Data)
	} else {
		fmt.Fprintln(os.Stderr, "get create to transaction error")
		return
	}
}

func wasmFuzzyQueryContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fuzzy query",
		Short: "fuzzy query the WASM contract for table's info within range",
		Run:   wasmFuzzyQueryContract,
	}
	wasmAddFuzzyQueryContractFlags(cmd)
	return cmd
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
	key, _ := cmd.Flags().GetString("key")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	queryReq := wasmtypes.WasmQueryContractTableReq{
		ContractName: contractName,
		Items:        []*wasmtypes.WasmQueryTableItem{{TableName: tableName, Key: key}},
	}

	var WasmQueryResponse wasmtypes.WasmQueryResponse
	query := sendQuery4wasm(rpcLaddr, wasmtypes.WasmGetContractTable, &queryReq, &WasmQueryResponse)
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

func wasmFuzzyQueryContract(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("exec")
	tableName, _ := cmd.Flags().GetString("table")
	format, _ := cmd.Flags().GetString("format")
	start, _ := cmd.Flags().GetInt64("start")
	stop, _ := cmd.Flags().GetInt64("stop")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	queryReq := wasmtypes.WasmFuzzyQueryTableReq{
		ContractName: contractName,
		TableName:    tableName,
		Format:       format,
		Start:        start,
		Stop:         stop,
	}

	var WasmQueryResponse wasmtypes.WasmFuzzyQueryResponse
	query := sendQuery4wasm(rpcLaddr, wasmtypes.WasmFuzzyGetContractTable, &queryReq, &WasmQueryResponse)
	if query {
		fmt.Println(WasmQueryResponse.ContractName)
		fmt.Println(WasmQueryResponse.TableName)
		for _, WasmOutItem := range WasmQueryResponse.QueryResultItems {
			fmt.Println(WasmOutItem.Index)
			for _, fuzzyQueryResultItem := range WasmOutItem.ResultJSON {
				fmt.Println(fuzzyQueryResultItem)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "get wasm query error")
		return
	}
}

func wasmAddFuzzyQueryContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "wasm contract name")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("table", "n", "", "one of wasm contract's table name")
	cmd.MarkFlagRequired("table")

	cmd.Flags().StringP("format", "f", "", "format of key's prefix for the table info")
	cmd.MarkFlagRequired("format")

	cmd.Flags().Int64P("start", "s", 0, "start value for the foramt")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Int64P("stop", "p", 0, "stop value for the foramt")
	cmd.MarkFlagRequired("stop")
}

func wasmAddQueryContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "wasm contract name")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("table", "n", "", "one of wasm contract's table name")
	cmd.MarkFlagRequired("table")

	cmd.Flags().StringP("key", "k", "", "key of the table info")
	cmd.MarkFlagRequired("key")
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

	var createReq = wasmtypes.CallContractReq{
		Name:       contractName,
		Note:       note,
		ActionName: actionName,
		DataInJson: abiPara,
		Fee:        int64(feeInt64),
	}
	var createResp = types.ReplyString{}

	query := sendQuery4wasm(rpcLaddr, wasmtypes.CallWasmContractStr, &createReq, &createResp)
	if query {
		fmt.Println(createResp.Data)
	} else {
		fmt.Fprintln(os.Stderr, "get call wasm to transaction error")
		return
	}
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


func addEstimateCallFlags(cmd *cobra.Command) {
	wasmAddCallContractFlags(cmd)

	cmd.Flags().StringP("caller", "c", "", "the caller address")

}
func wasmEstimateCallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "Estimate the gas cost of calling a contract",
		Run:   wasmestimatecall,
	}
	addEstimateCallFlags(cmd)
	return cmd
}

func wasmestimatecall(cmd *cobra.Command, args []string) {
	caller, _ := cmd.Flags().GetString("caller")

	fee, _ := cmd.Flags().GetFloat64("fee")
	contractName, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	abiPara, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var estGasResp wasmtypes.EstimateWASMGasResp

	feeInt64 := uint64(fee*1e4) * 1e4
	var estimatecallreq = wasmtypes.EstimateCallContractReq{
		Execer:     contractName,
		GasLimit:   feeInt64,
		From:       caller,
		ActionName: actionName,
		ActionData: []byte(abiPara),
	}
	query := sendQuery4wasm(rpcLaddr, "EstimateGasCallContract", &estimatecallreq, &estGasResp)

	if query {
		fmt.Fprintf(os.Stdout, "call contract gas cost estimate %v\n", estGasResp.Gas)
	} else {
		fmt.Fprintln(os.Stderr, "call contract gas cost estimate error")
	}

}

func addEstimateCreateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "d", "", "path where stores wasm code and abi")
	cmd.MarkFlagRequired("path")

	cmd.Flags().StringP("contract", "x", "", "contract name same with the code and abi file")
	cmd.MarkFlagRequired("contract")

}

func wasmestimatecreate(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	contractName, _ := cmd.Flags().GetString("contract")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nameReg, err := regexp.Compile(wasmtypes.NameRegExp)
	if !nameReg.MatchString(contractName) {
		fmt.Fprintln(os.Stderr, "Wrong wasm contract name format, which should be a-z and 0-9 ")
		return
	}

	if len(contractName) > 16 || len(contractName) < 4 {
		fmt.Fprintln(os.Stderr, "wasm contract name's length should be within range [4-16]")
		return
	}

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
	var estimatecreateReq = wasmtypes.EstimateCreateContractReq{
		Code: code,
		Abi:  string(abi),
	}
	var estGasResp wasmtypes.EstimateWASMGasResp
	query := sendQuery4wasm(rpcLaddr, "EstimateGasCreateContract", &estimatecreateReq, &estGasResp)
	if query {
		fmt.Fprintf(os.Stdout, "create contract gas cost estimate %v\n", estGasResp.Gas)
	} else {
		fmt.Fprintln(os.Stderr, "create contract gas cost estimate error")
		return
	}
}

func wasmEstimateCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Estimate the gas cost of creating a contract",
		Run:   wasmestimatecreate,
	}
	addEstimateCreateFlags(cmd)
	return cmd
}

// 估算合约消耗
func wasmEstimateContractCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate the gas cost of calling or creating a contract",
	}
	cmd.AddCommand(
		wasmEstimateCreateCmd(),
		wasmEstimateCallCmd())
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
	if bytes.Contains([]byte(name), []byte(wasmtypes.UserWasmX)) {
		name = name[len(wasmtypes.UserWasmX):]
	}

	match, _ := regexp.MatchString(wasmtypes.NameRegExp, name)
	if !match {
		fmt.Fprintln(os.Stderr, "Wrong wasm contract name format, which should be a-z and 0-9 ")
		return
	}

	var checkAddrReq = wasmtypes.CheckWASMContractNameReq{WasmContractName: name}
	var checkAddrResp wasmtypes.CheckWASMAddrResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4wasm(rpcLaddr, wasmtypes.CheckNameExistsFunc, &checkAddrReq, &checkAddrResp)
	if query {
		fmt.Fprintln(os.Stdout, checkAddrResp.ExistAlready)
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
