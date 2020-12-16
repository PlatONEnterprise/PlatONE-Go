#!/bin/bash

function help() {
    echo 
    echo "
USAGE: platonectl.sh [options]
    OPTIONS:
            --nodeid, -n                 the specified node id
            --content, -c                update content (default: '{"type":1}')
            --help, -h                   show help
"
}

function shiftOption2() {
    if [[ $1 -lt 2 ]];then
        echo "[ERROR]: ********* miss option value! please set the value **********"
        exit
    fi
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
P2P_PORT=16791
RPC_PORT=6791
nodeJsonStr_string='{"type":1}'
BOOTNODES=" "
IP=127.0.0.1
PUBKEY=""

CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts

while [ ! $# -eq 0 ]
do
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            ;;
        --content | -c)
            echo "updateContent: $2"
            nodeJsonStr_string=${2}
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
    echo "NODE_DIR: ${NODE_DIR}"
else 
    echo "[ERROR]: !!! There is no datadir of NODE_DIR !!!"
    exit 
fi

if [ -f ${NODE_DIR}/node.pubkey ]; then    
    PUBKEY=`cat ${NODE_DIR}/node.pubkey`
    echo "node's public key: ${PUBKEY}"
else
    echo "[ERROR]: !!! There is no datadir of NODE_DIR !!!"
    exit 
fi


config="${CONF_PATH}/ctool.json"
node_manager_abi="${CONF_PATH}/contracts/nodeManager.cpp.abi.json"

# get nodeManager address
#cns_manager_abi="${CONF_PATH}/contracts/cnsManager.cpp.abi.json"
#cns_addr="0x0000000000000000000000000000000000000011"
#func="getContractAddress"
#param1="__sys_NodeManager"
#param2="latest"
#ret=`${BIN_PATH}/ctool invoke --config $config --abi $cns_manager_abi --addr $cns_addr --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g`
#
#addr=${ret#*result:}
#echo "nodeManager's address: ${addr}"

name_string=${NODE_ID}

node_manager_addr="0x1000000000000000000000000000000000000002"

nodeJsonStr_string=`echo ${nodeJsonStr_string} | sed s/[[:space:]]//g`
${BIN_PATH}/ctool invoke --config $config --addr $node_manager_addr --abi $node_manager_abi --func update --param $name_string --param $nodeJsonStr_string

sleep 2

${BIN_PATH}/ctool invoke --config $config --addr $node_manager_addr --abi $node_manager_abi --func getAllNodes
