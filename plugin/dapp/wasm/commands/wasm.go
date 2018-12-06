package commands

//#cgo CPPFLAGS: -I./../../../wasmcpp/serialize -I./../../../wasmcpp/fc/include -I./../../../wasmcpp/adapter/include -I./../../../external/magic_get/include/
//#cgo LDFLAGS: -L/usr/local/lib -lboost_filesystem -lboost_system -lboost_chrono -lboost_date_time
//#cgo LDFLAGS: -L./../../../wasmcpp/lib -lwasm_serialize -lfc
//#include <../../../wasmcpp/serialize/serialize_api.h>
import "C"

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"
	"io/ioutil"

	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/rpc/jsonclient"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"

	"strconv"
	"encoding/json"
)

func WasmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "WASM contracts operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		wasmCreateContractCmd(),
		wasmGenAbiCmd(),
		wasmCallContractCmd(),
		wasmQueryContractCmd(),
		wasmEstimateContractCmd(),
		wasmCheckContractAddrCmd(),
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

	cmd.Flags().StringP("alias", "s", "", "human readable contract alias name(optional)")
}

func wasmCreateContract(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("contract")
	path, _ := cmd.Flags().GetString("path")

	//caller, _ := cmd.Flags().GetString("caller")
	//expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	alias, _ := cmd.Flags().GetString("alias")
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

	param := &wasmtypes.CreateOrCallWasmContract{
		Value: wasmtypes.CreateWasmContractPara{
			Code:code,
			Abi: string(abi),
			Alias: alias,
			Note: note,
			Fee: int64(feeInt64),
		},
	}

	paramJson, errInfo := json.Marshal(param)
	if errInfo != nil {
		fmt.Fprintln(os.Stderr, "json.Marshal error ", errInfo)
		return

	}

	params := &rpctypes.CreateTxIn{
		Execer:wasmtypes.WasmX,
		ActionName:"CreateCall",
		Payload:paramJson,
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
	contractAddr, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	abiPara, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	//从区块链获取abi文件，并根据abi文件将相应的json字符串转化为abi data
	req := types.ReqAddr{Addr: contractAddr}
	var res wasmtypes.WasmGetAbiResp
	var actionData []byte
	query := sendQuery4wasm(rpcLaddr, wasmtypes.WasmGetAbi, &req, &res)
	if query {
		actionData = genAbiData(string(res.Abi), contractAddr, actionName, abiPara)
		if nil == actionData {
			fmt.Fprintln(os.Stderr, "Failed to convert parameter from json to abi format")
			return
		}
	} else {
		fmt.Fprintln(os.Stderr, "get abi data error")
		return
	}

	queryReq := wasmtypes.WasmQuery{
		ContractAddr: contractAddr,
		ActionName:actionName,
		Abidata:actionData,
	}

	var WasmQueryResponse wasmtypes.WasmQueryResponse
	query = sendQuery4wasm(rpcLaddr, wasmtypes.QueryFromContract, &queryReq, &WasmQueryResponse)
	if query {
		for _, WasmOutItem := range WasmQueryResponse.QueryResultItems {
			fmt.Println(WasmOutItem.ItemType);
			fmt.Println(WasmOutItem.ResultJSON);
		}
	} else {
		fmt.Fprintln(os.Stderr, "get wasm query error")
		return
	}
}

func wasmAddQueryContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "wasm contract address")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("action", "x", "", "external contract action name")
	cmd.MarkFlagRequired("action")

	cmd.Flags().StringP("para", "r", "", "external contract execution parameter in json string")
	cmd.MarkFlagRequired("para")
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
	amount, _ := cmd.Flags().GetFloat64("amount")
	fee, _ := cmd.Flags().GetFloat64("fee")
	contractAddr, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	abiPara, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	amountInt64 := uint64(amount*1e4) * 1e4
	feeInt64 := uint64(fee*1e4) * 1e4

	//从区块链获取abi文件，并根据abi文件将相应的json字符串转化为abi data
	req := types.ReqAddr{Addr: contractAddr}
	var res wasmtypes.WasmGetAbiResp
	var actionData []byte
	query := sendQuery4wasm(rpcLaddr, "WasmGetAbi", &req, &res)
	if query {
		actionData = genAbiData(string(res.Abi), contractAddr, actionName, abiPara)
		//fmt.Println(string("The converted abi data is:") + common.ToHex(actionData))
	} else {
		fmt.Fprintln(os.Stderr, "get abi data error")
		return
	}

	param := &wasmtypes.CreateOrCallWasmContract{
		Value: wasmtypes.CallWasmContractPara{
			Amount:amountInt64,
			ContractAddr: contractAddr,
			Alias:"",
			Note: note,
			ActionName:actionName,
			ActionData:actionData,
			Fee: int64(feeInt64),
		},
	}

	paramJson, errInfo := json.Marshal(param)
	if errInfo != nil {
		fmt.Fprintln(os.Stderr, "json.Marshal error ", errInfo)
		return

	}

	params := &rpctypes.CreateTxIn{
		Execer:wasmtypes.WasmX,
		ActionName:"CreateCall",
		Payload:paramJson,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func wasmAddCallContractFlags(cmd *cobra.Command) {
	wasmAddCommonFlags(cmd)
	cmd.Flags().StringP("exec", "e", "", "wasm contract address")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("action", "x", "", "external contract action name")
	cmd.MarkFlagRequired("action")

	cmd.Flags().StringP("para", "r", "", "external contract execution parameter in json string")
	cmd.MarkFlagRequired("para")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")
}

func wasmAddCommonFlags(cmd *cobra.Command) {
	//cmd.Flags().StringP("caller", "c", "", "the caller address")
	//cmd.MarkFlagRequired("caller")

	//cmd.Flags().StringP("expire", "p", "120s", "transaction expire time (optional)")

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
func wasmCheckContractAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if the address is a valid WASM contract",
		Run:   wasmCheckContractAddr,
	}
	wasmAddCheckContractAddrFlags(cmd)
	return cmd
}

func wasmAddCheckContractAddrFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "external contract address (optional)")
	cmd.Flags().StringP("exec", "e", "", "external contract name, like user.external.xxxxx (optional)")
}

func wasmCheckContractAddr(cmd *cobra.Command, args []string) {
	to, _ := cmd.Flags().GetString("to")
	name, _ := cmd.Flags().GetString("exec")
	toAddr := to
	if len(toAddr) == 0 && len(name) > 0 {
		if strings.Contains(name, wasmtypes.UserWasmX) {
			toAddr = address.ExecAddress(name)
		}
	}
	if len(toAddr) == 0 {
		fmt.Fprintln(os.Stderr, "one of the 'to (contract address)' and 'name (contract name)' must be set")
		cmd.Help()
		return
	}

	var checkAddrReq = wasmtypes.CheckWASMAddrReq{Addr: toAddr}
	var checkAddrResp wasmtypes.CheckWASMAddrResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4wasm(rpcLaddr, "CheckAddrExists", &checkAddrReq, &checkAddrResp)

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
