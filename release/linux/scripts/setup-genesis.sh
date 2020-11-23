#!/bin/bash

########################################

function shiftOption2() {
    if [[ $1 -lt 2 ]];then
        echo "[ERROR]: ********* miss option value! please set the value **********"
        exit
    fi
}

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

function create_node_key() {
    keyinfo=`${BIN_PATH}/ethkey genkeypair | sed s/[[:space:]]//g`
    keyinfo=${keyinfo,,}
    address=${keyinfo:10:40}
    prikey=${keyinfo:62:64}
    pubkey=${keyinfo:137:128}

    if [ ${#prikey} -ne 64 ]; then
        echo "Error: create node key failed."
        exit
    fi

    mkdir -p ${NODE_DIR}

    ts=`date '+%Y%m%d%H%M%S'`
    if [ -f ${NODE_DIR}/node.address ]; then
        mkdir -p ${NODE_DIR}/bak
        mv ${NODE_DIR}/node.address ${NODE_DIR}/bak/node.address.bak.$ts
    fi
    if [ -f ${NODE_DIR}/node.prikey ]; then
        mkdir -p ${NODE_DIR}/bak
        mv ${NODE_DIR}/node.prikey ${NODE_DIR}/bak/node.prikey.bak.$ts
    fi
    if [ -f ${NODE_DIR}/node.pubkey ]; then
        mkdir -p ${NODE_DIR}/bak
        mv ${NODE_DIR}/node.pubkey ${NODE_DIR}/bak/node.pubkey.bak.$ts
    fi
    
    echo "node's address: $address"
    echo $address > ${NODE_DIR}/node.address
    echo "node's private key: $prikey"  
    echo $prikey > ${NODE_DIR}/node.prikey
    echo "node's public key: $pubkey"
    echo $pubkey > ${NODE_DIR}/node.pubkey

    echo "Create node key succ. Files: ${NODE_DIR}/node.address, ${NODE_DIR}/node.prikey, ${NODE_DIR}/node.pubkey"
}

function replaceList() {
    res=`echo $2 | sed "s/,/\",\"/g"`
    ${BIN_PATH}/repstr ${CONF_PATH}/genesis.json $1 `echo \"${res}\"`
}

function create_genesis() {
    if [ -f ${CONF_PATH}/genesis.json ]; then
        mkdir -p ${CONF_PATH}/bak
        mv ${CONF_PATH}/genesis.json ${CONF_PATH}/bak/genesis.json.bak.`date '+%Y%m%d%H%M%S'`
    fi
    cp ${CONF_PATH}/genesis.json.istanbul.template ${CONF_PATH}/genesis.json


    NODE_KEY=`cat ${NODE_DIR}/node.pubkey`
    default_enode="enode://${NODE_KEY}@${1}:${2}"
    if [[ $VALIDATOR_NODES != "" ]]; then
         replaceList "__VALIDATOR__" $VALIDATOR_NODES
    else
         replaceList "__VALIDATOR__" $default_enode
    fi

     ${BIN_PATH}/repstr ${CONF_PATH}/genesis.json "DEFAULT-ACCOUNT" 0000000000000000000000000000000000000001
     ${BIN_PATH}/repstr ${CONF_PATH}/genesis.json "__INTERPRETER__" ${INTERPRETER}

#    ${BIN_PATH}/ctool codegen --abi ${CONF_PATH}/contracts/cnsProxy.cpp.abi.json --code ${CONF_PATH}/contracts/cnsProxy.wasm > ${CONF_PATH}/cns-code.hex
    
#    ${BIN_PATH}/repstr ${CONF_PATH}/genesis.json "CNS-CODE" -f ${CONF_PATH}/cns-code.hex
#    rm -rf ${CONF_PATH}/cns-code.hex

    now=`date +%s`

    ${BIN_PATH}/repstr ${CONF_PATH}/genesis.json "TIMESTAMP" $now

    echo "[INFO]: Create genesis succ. File: ${CONF_PATH}/genesis.json"
}

function setup_genesis() {
    if [ "${AUTO}" = "true" ]; then 
        echo "[INFO]: auto create node key, and create genesis.json"
        create_node_key
        echo $IP > ${NODE_DIR}/node.ip
        echo; echo "[Create genesis]"
        create_genesis $IP ${P2P_PORT}
        return
    else
        echo "AUTO: ${AUTO}"  
    fi

    ## 1. create node key
    echo; echo "Do You What To Create a new node key ?"
    yes_or_no
    if [ $? -eq 1 ]; then        
        if [ -f ${NODE_DIR}/node.pubkey ]; then
            echo "[INFO]: Node key already exists, re create?"
            yes_or_no
            if [ $? -eq 1 ]; then
                create_node_key
            fi
        else
            create_node_key
        fi
    else
        if [ -f ${NODE_DIR}/node.pubkey ]; then
            pubkey=`cat ${NODE_DIR}/node.pubkey`
            echo "Node's public key: ${pubkey}"
        else
            echo; echo "!!! No node's public key file !!!"
            echo "[WARN]: Please Put Your Nodekey file \"node.pubkey\" to the directory ${NODE_DIR}"
            exit
        fi
    fi


    ## 2. Input public ip addr
    echo "first node ip: ${IP}"
    check_ip $IP
    if [ $? -eq 0 ]; then
        echo $IP > ${NODE_DIR}/node.ip
    else
        echo "[WARN]: Invalid ip! Please input a valid ip address."
        echo; echo "[Input public ip addr]"
        while true
        do
            read -p "Your node ip: " ip_input 
            check_ip $ip_input
            if [ $? -eq 0 ]; then
                echo $ip_input > ${NODE_DIR}/node.ip
                break
            else
                echo "[ERROR]: Invalid ip. Please re input."
            fi
        done
    fi

    IP=`cat ${NODE_DIR}/node.ip`

    echo; echo "[Create genesis]"
    create_genesis $IP ${P2P_PORT}
}

function compile_system_contracts() {
    if [ "${AUTO}" = "true" ]; then 
        echo "[INFO]: auto skip compile sys contract, Use existing"
        return
    else
        echo "AUTO: ${AUTO}"  
    fi

    # recompile the system contracts
    echo "[INFO]: Do You What To Recompile The System Contracts ? (Make sure to put the source code of the system contract in ${SYS_CONTRACTS_PATH})"
    yes_or_no
    if [ $? -eq 0 ]; then
        return
    fi

    if [[ -d ${SYS_CONTRACTS_PATH} ]]; then
        # Recompile system contract
        cd ${SYS_CONTRACTS_PATH}

        rm -rf ${SYS_CONTRACTS_PATH}/build
        ./script/build_system_contracts.sh .
        cp ${SYS_CONTRACTS_PATH}/build/systemContract/*/*json ${SYS_CONTRACTS_PATH}/build/systemContract/*/*wasm  ${WORKSPACE_PATH}/conf/contracts

        cd ${CURRENT_PATH}
    else
        echo "[ERROR]: not found the source code of the system contract; check the source code path: ${SYS_CONTRACTS_PATH}"
    fi
}


###########################################
#### Setup the genesis.json of a chain ####
###########################################

function help() {
    echo 
    echo "
USAGE: platonectl.sh setupgen [options]

        OPTIONS:
           --nodeid, -n                 the first node id (default: 0)
           --ip                         the first node ip (default: 127.0.0.1)
           --p2p_port                   the first node p2p_port (default: 16791)
           --auto                       auto=true: Will auto create new node keys and will
                                        not compile system contracts again (default=false)
           --observerNodes, -o          set the genesis suggestObserverNodes
                                        (default is the first node enode code)
           --validatorNodes, -v         set the genesis validatorNodes
                                        (default is the first node enode code)
           --help, -h                   show help
"
}

# variables
CURRENT_PATH=`pwd`
cd ${CURRENT_PATH}
NODE_ID=0
IP=127.0.0.1
P2P_PORT=16791
OBSERVE_NODES=""
VALIDATOR_NODES=""
AUTO=false
INTERPRETER="wasm"

CURRENT_PATH=`pwd`
cd ${CURRENT_PATH}/..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts

if [[ ! -d ${WORKSPACE_PATH}/conf/contracts ]];then
    echo "[INFO]: create contracts dir in: ${WORKSPACE_PATH}/conf/contracts"
    echo "[WARN]: Please compile the system contract next, will auto put the bytecode and abi in this directory: ${WORKSPACE_PATH}/conf/contracts"
    mkdir ${WORKSPACE_PATH}/conf/contracts
fi

SYS_CONTRACTS_PATH=${WORKSPACE_PATH}/../../cmd/SysContracts

while [ ! $# -eq 0 ]
do
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            ;;
        --ip)
            echo "ip: $2"
            IP=$2
            ;;
        --p2p_port | -p)
            echo "p2p_port: $2"
            P2P_PORT=$2
            ;;
        --auto)
            echo "auto: $2"
            AUTO=$2
            ;;
        --observerNodes | -o)
            echo "bootnodes: $2"
            OBSERVE_NODES=$2
            ;;
        --validatorNodes | -v)
            echo "bootnodes: $2"
            VALIDATOR_NODES=$2
            ;;
        --interpreter | -i)
            echo "interpreter: #${2}#"
            INTERPRETER=${2}
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

if [ -d ${NODE_DIR} ]; then
    echo "root node datadir: ${NODE_DIR}"
else
    echo '[INFO]: The node directory have not been created, Now to create it'
    mkdir -p ${NODE_DIR}
fi

echo ${P2P_PORT} > ${NODE_DIR}/node.p2p_port

echo '
###########################################
#### Setup the genesis.json of a chain ####
###########################################
'

compile_system_contracts

setup_genesis
