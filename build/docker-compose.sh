#!/usr/bin/env bash

set -e
set -o pipefail
#set -o verbose
#set -o xtrace

# os: ubuntu16.04 x64
# first, you must install jq tool of json
# sudo apt-get install jq
# sudo apt-get install shellcheck, in order to static check shell script
# sudo apt-get install parallel
# ./docker-compose.sh build

PWD=$(cd "$(dirname "$0")" && pwd)
export PATH="$PWD:$PATH"

NODE3="${1}_chain33_1"
CLI="docker exec ${NODE3} /root/chain33-cli"
PARA_CLI="docker exec ${NODE3} /root/chain33-para-cli"

NODE2="${1}_chain32_1"

NODE1="${1}_chain31_1"

NODE4="${1}_chain30_1"
#CLI4="docker exec ${NODE4} /root/chain33-cli"

NODE5="${1}_chain29_1"
CLI5="docker exec ${NODE5} /root/chain33-cli"

containers=("${NODE1}" "${NODE2}" "${NODE3}" "${NODE4}")
export COMPOSE_PROJECT_NAME="$1"
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

DAPP=""
if [ -n "${2}" ]; then
    DAPP=$2
fi

DAPP_TEST_FILE=""
if [ -n "${DAPP}" ]; then
    DAPP_TEST_FILE="testcase.sh"
    if [ -e "$DAPP_TEST_FILE" ]; then
        # shellcheck source=/dev/null
        source "${DAPP_TEST_FILE}"
    fi

    DAPP_COMPOSE_FILE="docker-compose-${DAPP}.yml"
    if [ -e "$DAPP_COMPOSE_FILE" ]; then
        export COMPOSE_FILE="docker-compose.yml:${DAPP_COMPOSE_FILE}"

    fi

fi

echo "=========== # env setting ============="
echo "DAPP=$DAPP"
echo "DAPP_TEST_FILE=$DAPP_TEST_FILE"
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "CLI=$CLI"
####################

function base_init() {

    # update test environment
    sed -i $sedfix 's/^Title.*/Title="local"/g' chain33.toml
    sed -i $sedfix 's/^TestNet=.*/TestNet=true/g' chain33.toml

    # p2p
    sed -i $sedfix 's/^seeds=.*/seeds=["chain33:13802","chain32:13802","chain31:13802"]/g' chain33.toml
    #sed -i $sedfix 's/^enable=.*/enable=true/g' chain33.toml
    sed -i $sedfix '0,/^enable=.*/s//enable=true/' chain33.toml
    sed -i $sedfix 's/^isSeed=.*/isSeed=true/g' chain33.toml
    sed -i $sedfix 's/^innerSeedEnable=.*/innerSeedEnable=false/g' chain33.toml
    sed -i $sedfix 's/^useGithub=.*/useGithub=false/g' chain33.toml

    # rpc
    sed -i $sedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8801"/g' chain33.toml
    sed -i $sedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8802"/g' chain33.toml
    sed -i $sedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' chain33.toml

    # wallet
    sed -i $sedfix 's/^minerdisable=.*/minerdisable=false/g' chain33.toml

}

function start() {
    echo "=========== # docker-compose ps ============="
    docker-compose ps

    # remove exsit container
    docker-compose down

    # create and run docker-compose container
    #docker-compose -f docker-compose.yml -f docker-compose-paracross.yml -f docker-compose-relay.yml up --build -d
    docker-compose up --build -d

    local SLEEP=10
    echo "=========== sleep ${SLEEP}s ============="
    sleep ${SLEEP}

    docker-compose ps

    # query node run status
    check_docker_status
    ${CLI} block last_header
    ${CLI} net info

    ${CLI} net peer_info
    local count=100
    while [ $count -gt 0 ]; do
        peersCount=$(${CLI} net peer_info | jq '.[] | length')
        if [ "${peersCount}" -ge 2 ]; then
            break
        fi
        sleep 5
        ((count--))
        echo "peers error: peersCount=${peersCount}"
    done

    miner "${CLI}"
    block_wait "${CLI}" 1

    echo "=========== check genesis hash ========== "
    ${CLI} block hash -t 0
    res=$(${CLI} block hash -t 0 | jq ".hash")
    count=$(echo "$res" | grep -c "0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6")
    if [ "${count}" != 1 ]; then
        echo "genesis hash error!"
        exit 1
    fi

    echo "=========== query height ========== "
    ${CLI} block last_header
    result=$(${CLI} block last_header | jq ".height")
    if [ "${result}" -lt 1 ]; then
        block_wait "${CLI}" 2
    fi

    sync_status "${CLI}"

    ${CLI} wallet status
    ${CLI} account list
    ${CLI} mempool list
}

function miner() {
    #echo "=========== # create seed for wallet ============="
    #seed=$(${1} seed generate -l 0 | jq ".seed")
    #if [ -z "${seed}" ]; then
    #    exit 1
    #fi

    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314 -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    sleep 1

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314 -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    sleep 1

    echo "=========== # import private key returnAddr ============="
    result=$(${1} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    sleep 1

    echo "=========== # import private key mining ============="
    result=$(${1} account import_key -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01 -l minerAddr | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    sleep 1
    echo "=========== # close auto mining ============="
    result=$(${1} wallet auto_mine -f 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

}
function block_wait() {
    if [ "$#" -lt 2 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    cur_height=$(${1} block last_header | jq ".height")
    expect=$((cur_height + ${2}))
    local count=0
    while true; do
        new_height=$(${1} block last_header | jq ".height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

function check_docker_status() {
    status=$(docker-compose ps | grep chain33_1 | awk '{print $6}')
    statusPara=$(docker-compose ps | grep chain33_1 | awk '{print $3}')
    if [ "${status}" == "Exit" ] || [ "${statusPara}" == "Exit" ]; then
        echo "=========== chain33 service Exit logs ========== "
        docker-compose logs chain33
        echo "=========== chain33 service Exit logs End========== "
    fi

}

function check_docker_container() {
    echo "============== check_docker_container ==============================="
    for con in "${containers[@]}"; do
        runing=$(docker inspect "${con}" | jq '.[0].State.Running')
        if [ ! "${runing}" ]; then
            docker inspect "${con}"
            echo "check ${con} not actived!"
            exit 1
        fi
    done
}

function sync_status() {
    echo "=========== query sync status========== "
    local sync_status
    local count=100
    local wait_sec=0
    while [ $count -gt 0 ]; do
        sync_status=$(${1} net is_sync)
        if [ "${sync_status}" = "true" ]; then
            break
        fi
        ((count--))
        wait_sec=$((wait_sec + 1))
        sleep 1
    done
    echo "sync wait  ${wait_sec} s"

    echo "=========== query clock sync status========== "
    sync_status=$(${1} net is_clock_sync)
    if [ "${sync_status}" = "false" ]; then
        exit 1
    fi
}

function sync() {
    echo "=========== stop  ${NODE5} node========== "
    docker stop "${NODE5}"
    sleep 10

    echo "=========== start ${NODE5} node========== "
    docker start "${NODE5}"

    sleep 1
    sync_status "${CLI5}"
}

function transfer() {
    echo "=========== # transfer ============="
    hashes=()
    for ((i = 0; i < 10; i++)); do
        hash=$(${CLI} send coins transfer -a 1 -n test -t 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
        hashes=("${hashes[@]}" "$hash")
    done
    block_wait "${CLI}" 1
    echo "len: ${#hashes[@]}"
    if [ "${#hashes[@]}" != 10 ]; then
        echo "tx number wrong"
        exit 1
    fi

    for ((i = 0; i < ${#hashes[*]}; i++)); do
        txs=$(${CLI} tx query_hash -s "${hashes[$i]}" | jq ".txs")
        if [ -z "${txs}" ]; then
            echo "cannot find tx"
            exit 1
        fi
    done

    echo "=========== # withdraw ============="
    hash=$(${CLI} send coins transfer -a 2 -n deposit -t 1wvmD6RNHzwhY4eN75WnM6JcaAvNQ4nHx -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    block_wait "${CLI}" 1
    before=$(${CLI} account balance -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e retrieve | jq -r ".balance")
    if [ "${before}" == "0.0000" ]; then
        echo "wrong ticket balance, should not be zero"
        exit 1
    fi

    hash=$(${CLI} send coins withdraw -a 1 -n withdraw -e retrieve -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    block_wait "${CLI}" 1
    txs=$(${CLI} tx query_hash -s "${hash}" | jq ".txs")
    if [ "${txs}" == "null" ]; then
        echo "withdraw cannot find tx"
        exit 1
    fi
}

function dice_test() {

    echo "== unlock wallet =="
    result=`${PARA_CLI} wallet unlock -p 1314 | jq '.isOK'`
    if [[ ${result} != "true" ]]; then
        echo "wallet unlock error"
        exit 1
    fi

    echo "para import private key: CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
    ${PARA_CLI} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr

    echo "para import private key: 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
    ${PARA_CLI} account import_key -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01 -l minerAddr

    echo "== paracross transfer bty =="
    hash=`${PARA_CLI} send bty transfer -a 10000 -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "transfer bty failed"
        exit 1
    fi

    echo "== create dice contract =="
    hash=`${PARA_CLI} send wasm create -x dice -f 10 -n deployDice -d . -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "create dice contract failed"
        exit 1
    fi

    echo "== transfer bty to dice =="
    ${PARA_CLI} send bty send_exec -e user.p.para.user.wasm.dice -a 500 -n transfer2dice -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    hash=`${PARA_CLI} send bty send_exec -e user.p.para.user.wasm.dice -a 5000 -n transfer2dice -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "transfer bty to dice failed"
        exit 1
    fi

    echo "== startgame =="
    hash=`${PARA_CLI} send wasm call -x startgame -e user.p.para.user.wasm.dice -f 0.002 -n start -r "{\"deposit\":\"4000\"}" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "start game failed"
        exit 1
    fi

    echo "== query game status =="
    result=`${PARA_CLI} wasm query -e user.p.para.user.wasm.dice -k dice_statics -n gamestatus`
    echo ${result}
    balance=`echo ${result} | jq '.game_balance'`
    if [[ ${balance} != 400000 ]]; then
        echo "query game failed"
        exit 1
    fi
    active=`echo ${result} | jq '.is_active'`
    if [[ ${active} != 1 ]]; then
        echo "query game failed"
        exit 1
    fi
    creator=`echo ${result} | jq '.game_creator'`
    if [[ ${creator} != '"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"' ]]; then
        echo "query game failed"
        exit 1
    fi

    echo "== play game =="
    hash=`${PARA_CLI} send wasm call -x play -e user.p.para.user.wasm.dice -f 0.002 -r "{\"amount\":2,\"number\":30,\"direction\":0}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "play game failed"
        exit 1
    fi

    echo "== query round info =="
    result=`${PARA_CLI} wasm query -e user.p.para.user.wasm.dice -k round:1 -n roundinfo`
    echo ${result}
    num=`echo ${result} | jq '.guess_num'`
    if [[ ${num} != 30 ]]; then
        echo "query round info failed"
        exit 1
    fi
    round=`echo ${result} | jq '.round'`
    if [[ ${round} != 1 ]]; then
        echo "query round info failed"
        exit 1
    fi
    amount=`echo ${result} | jq '.amount'`
    if [[ ${amount} != 2 ]]; then
        echo "query round info failed"
        exit 1
    fi
    player=`echo ${result} | jq '.player'`
    if [[ ${player} != '"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"' ]]; then
        echo "query round info failed"
        exit 1
    fi

    echo "== stop game =="
    hash=`${PARA_CLI} send wasm call -x stopgame -e user.p.para.user.wasm.dice -f 0.002 -r "{}" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
    echo "${hash}"
    block_wait "${PARA_CLI}" 2
    result=`${PARA_CLI} tx query -s ${hash} | jq '.receipt.tyName'`
    if [[ ${result} != '"ExecOk"' ]]; then
        echo "stop game failed"
        exit 1
    fi

    echo "== query game status =="
    result=`${PARA_CLI} wasm query -e user.p.para.user.wasm.dice -k dice_statics -n gamestatus`
    echo ${result}
    active=`echo ${result} | jq '.is_active'`
    if [[ ${active} != 0 ]]; then
        echo "query game failed"
        exit 1
    fi

    echo "== check balance =="
    result=`${PARA_CLI} account balance -e user.p.para.user.wasm.dice -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt | jq '.frozen'`
    if [[ ${result} != '"0.0000"' ]]; then
        echo "check balance failed"
        exit 1
    fi

    echo "dice contract test ok!"
    echo ""
}

function base_config() {
    sync
    transfer
}

function dapp_run() {
    if [ -e "$DAPP_TEST_FILE" ]; then
        ${DAPP} "${CLI}" "${1}"
    fi

}
function main() {
    echo "==============================DAPP=$DAPP main begin========================================================"
    ### init para ####
    base_init
    dapp_run init

    ### start docker ####
    start

    ### config env ###
    base_config
    dapp_run config

    ### test cases ###
    dapp_run test

    ### finish ###
    check_docker_container

    ### test wasm dice ###
    if [ -n "${DAPP}" ]; then
        dice_test
    fi
    echo "===============================DAPP=$DAPP main end========================================================="
}

# run script
main
