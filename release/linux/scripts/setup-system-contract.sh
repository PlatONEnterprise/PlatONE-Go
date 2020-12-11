#!/bin/bash

function yes_or_no() {
    read -p "Yes or No(y/n): " anw
    case $anw in
    [Yy][Ee][Ss]|[yY])
        return 1
    ;;
    [Nn][Oo]|[Nn])
        return 0
    ;;
    esac
    return 0
}

function shiftOption2() {
    if [[ $1 -lt 2 ]];then
        echo "[ERROR]: ********* miss option value! please set the value **********"
        exit
    fi
}

function check_ip() {
    ip=$1
    check=$(echo $ip|awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo $ip|grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$">/dev/null; then
        if [ ${check:-no} == "yes" ]; then
            return 0
        fi
    fi
    return 1
}

function create_account() {
    phrase=0

    if [ "${AUTO}" = "true" ]; then
        echo "[INFO]: auto use default password 0 to create the account"
    else
        echo "[INFO]: Input account passphrase."
        read -p "passphrase: " phrase
    fi
    echo "$IP:$RPC_PORT"
    ret=$( curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${phrase}\"],\"id\":1}"  http://${IP}:${RPC_PORT} )
    echo $ret

    substr=${ret##*\"result\":\"}

    if [ ${#substr} -gt 42 ]; then
        ACCOUNT=${substr:0:42}
        echo "New account: "${ACCOUNT}
        echo "passphrase: "${phrase}
        unlock_account ${ACCOUNT} ${phrase}
        cp $DATA_DIR/node-0/keystore/UTC* $CONF_PATH/keyfile.json
    else
        echo "[ERROR]: create account failed!!!"
        exit
    fi
}

function unlock_account() {
    http_data="{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"$1\",\"$2\",0],\"id\":1}"
    echo $http_data
    curl -H "Content-Type: application/json" --data "${http_data}"  http://${IP}:${RPC_PORT}
}

function create_ctooljson() {
    if [ -f ${CONF_PATH}/ctool.json ]; then
        mkdir -p ${CONF_PATH}/bak
        mv ${CONF_PATH}/ctool.json ${CONF_PATH}/bak/ctool.json.bak.`date '+%Y%m%d%H%M%S'`
    fi
    cp ${CONF_PATH}/ctool.json.template ${CONF_PATH}/ctool.json

    ${BIN_PATH}/repstr ${CONF_PATH}/ctool.json "NODE-IP" ${IP}
    ${BIN_PATH}/repstr ${CONF_PATH}/ctool.json "RPC-PORT" ${RPC_PORT}
    echo "${IP}:${RPC_PORT}"

    ${BIN_PATH}/repstr ${CONF_PATH}/ctool.json "DEFAULT-ACCOUNT" ${ACCOUNT:2}
    echo ${ACCOUNT:2}

    echo "[INFO]: Create ctool.json succ. File: ${CONF_PATH}/ctool.json"
}

function deploy() {
    echo ""
    echo "[INFO]: ******* to deploy system contract ${1} ******"
    name=$1
    config="${CONF_PATH}/ctool.json"
    code="${CONF_PATH}/contracts/$name.wasm"
    abi="${CONF_PATH}/contracts/$name.cpp.abi.json"

    ret=`${BIN_PATH}/ctool deploy --config $config --code $code --abi $abi | sed s/[[:space:]]//g`

    address=${ret#*contractaddress:}
    if [ ${#address} -eq 42 ]; then
        echo "${ret}"
        echo "[INFO]: $name deployed succ. Address: $address"
    else
        echo "[ERROR]: $name deployed failed."
    fi
}

function readEnv() {
    if [ -f ${NODE_DIR}/node.ip ]; then
        IP=`cat ${NODE_DIR}/node.ip`
        echo "node.ip: ${IP}"
    else
        echo 'The node.ip have not been created'
        exit
    fi

    if [ -f ${NODE_DIR}/node.rpc_port ]; then
        RPC_PORT=`cat ${NODE_DIR}/node.rpc_port`
        echo "node.rpc: ${RPC_PORT}"
    else
        echo 'The node.ip have not been created'
        exit
    fi

    if [ -f ${NODE_DIR}/node.p2p_port ]; then
        P2P_PORT=`cat ${NODE_DIR}/node.p2p_port`
        echo "node.p2p: ${P2P_PORT}"
    else
        echo 'The node.ip have not been created'
        exit
    fi
}

function add_first_node() {
    echo "[INFO]: =========================add first node id: ${NODE_ID} to sys contract=========================="
    ${SCRIPT_PATH}/add-node.sh --nodeid $NODE_ID
}

function set_super_admin(){
    ${BIN_PATH}/ctool invoke --config ${CONF_PATH}/ctool.json --abi ${CONF_PATH}/contracts/userManager.cpp.abi.json --addr 0x1000000000000000000000000000000000000001 --func setSuperAdmin
}

function add_chain_admin(){
    ${BIN_PATH}/ctool invoke --config ${CONF_PATH}/ctool.json --abi ${CONF_PATH}/contracts/userManager.cpp.abi.json --addr 0x1000000000000000000000000000000000000001 --func addChainAdminByAddress --param ${ACCOUNT}
}

function add_ca_certs(){

    if [ $ORG_CERT=="" ]; then 
        ORG_CERT=${CA_PATH}/org.crt
    fi 

    if [ $ROOT_CERT=="" ]; then 
        ROOT_CERT=${CA_PATH}/root.crt
    fi 

    root=`cat $ROOT_CERT`
    org=`cat $ORG_CERT`

    echo $root 
    echo $org

    echo `
    
    `

    ${BIN_PATH}/platonecli ca setRootCert --ca ${ROOT_CERT} --keyfile ${CONF_PATH}/keyfile.json 
    ${BIN_PATH}/platonecli ca addIssuer --ca ${ORG_CERT} --keyfile ${CONF_PATH}/keyfile.json 
}

function main() {
    readEnv

    echo "[INFO] An account will be created. If auto=true is set, the default password is 0"
    create_account
    echo "[INFO] to create ctool.json"
    create_ctooljson

    #deploy cnsManager
    #deploy paramManager
    #deploy userManager
    #deploy userRegister
    #deploy roleManager
    #deploy roleRegister
    #deploy nodeManager
#    deploy nodeRegister

    set_super_admin
    add_chain_admin

    add_first_node

    add_ca_certs

    echo "[INFO]: ========================= update first node id: ${NODE_ID} to consensus node =========================="
    ./update_to_consensus_node.sh -n $NODE_ID
}

function help() {
    echo
    echo "
USAGE: platonectl.sh deploysys [options]

        OPTIONS:
            --nodeid, -n                 the specified node id (default: 0)
            --root_cert                  root cert
            --node_cert                  node cert
            --auto                       auto=true: will use the default node password: 0
                                         to create the account and also
                                         to unlock the account. (default: false)
            --help, -h                   show help
"
}

# variables
SCRIPT_NAME=$0
SCRIPT_DIR_R=`dirname "${SCRIPT_NAME}"`
CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR_R}
SCRIPT_DIR=`pwd`
cd ${CURRENT_PATH}
NODE_ID=0
#SCRIPT_DIR="$( cd -P "$( dirname "${SCRIPT_NAME}" )" >/dev/null 2>&1 && pwd )"NODE_ID=0
IP=127.0.0.1
RPC_PORT=6791
P2P_PORT=16791
BOOTNODES=" "
AUTO=false

ACCOUNT=""

CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts
DATA_DIR=${WORKSPACE_PATH}/data
CA_PATH=${WORKSPACE_PATH}/ca-certs

ROOT_CERT=""
ORG_CERT=""

while [ ! $# -eq 0 ]
do
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            ;;
        --root_cert)
            echo "root cert: $2"
            ROOT_CERT=$2
            ;;
        --org_cert)
            echo "org cert: $2"
            ORG_CERT=$2
            ;;
        --auto)
            AUTO=$2
            ;;
        *)
            help
            exit
            ;;
    esac
    shiftOption2 $#
    shift 2
done

NODE_DIR=${WORKSPACE_PATH}/data/node-${NODE_ID}

main
