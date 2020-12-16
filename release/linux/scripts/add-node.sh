#!/bin/bash

function help() {
    echo 
    echo "
USAGE: platonectl.sh addnode [options] [value]

        OPTIONS:

           --nodeid, -n                 the specified node id. must be specified
           --desc                       the specified node desc
           --p2p_port                   the specified node p2p_port
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.
           --rpc_port                   the specified node rpc_port
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.
           --ip                         the specified node ip
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.
           --pubkey                     the specified node pubkey
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.
           --account                    the specified node account
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.
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
CURRENT_PATH=`pwd`
cd ${CURRENT_PATH}
NODE_ID=0

NODE_TYPE=0
P2P_PORT=16791
RPC_PORT=6791
BOOTNODES=" "
IP=127.0.0.1
PUBKEY=""
IS_ROOT=false
ACCOUNT=""
DESC=""
CURRENT_PATH=`pwd`
cd ${CURRENT_PATH}/..
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
        --account)
            echo "account: $2"
            account=$2
            ;;
        --desc)
            echo "desc: $2"
            DESC=$2
            ;;
        --p2p_port)
            echo "p2p_port: $2"
            p2p_port=$2
            ;;
        --rpc_port)
            echo "p2p_port: $2"
            rpc_port=$2
            ;;
        --ip)
            echo "ip: $2"
            ip=$2
            ;;
        --pubkey)
            echo "public key: $2"
            pubkey=$2
            ;;
        *)
            help
            exit
            ;;
    esac
    shiftOption2 $#
    shift 2
done



function readEnv() {
    if [ -d ${NODE_DIR} ]; then
        echo "NODE_DIR: ${NODE_DIR}"
    else
        echo "[INFO]: no datadir local. read remote node!!!"
        return
    fi

    if [ -f ${NODE_DIR}/node.ip ]; then
        IP=`cat ${NODE_DIR}/node.ip`
        echo "node.ip: ${IP}"
    else
        echo '[ERROR]: The node.ip have not been created'
        exit
    fi

    if [ -f ${NODE_DIR}/node.rpc_port ]; then
        RPC_PORT=`cat ${NODE_DIR}/node.rpc_port`
        echo "node.rpc: ${RPC_PORT}"
    else
        echo '[ERROR]: The node.rpc_port have not been created'
        exit
    fi

    if [ -f ${NODE_DIR}/node.p2p_port ]; then
        P2P_PORT=`cat ${NODE_DIR}/node.p2p_port`
        echo "node.p2p: ${P2P_PORT}"
    else
        echo '[ERROR]: !!! There is no datadir of p2p_port '
        exit
    fi

    if [ -f ${NODE_DIR}/node.pubkey ]; then
        PUBKEY=`cat ${NODE_DIR}/node.pubkey`
        echo "node's public key: ${PUBKEY}"
    else
        echo "[ERROR]: !!! There is no datadir of pubkey !!!"
        exit
    fi

    if [ -d ${NODE_DIR}/keystore ]; then
        keystore=${NODE_DIR}/keystore
        keys=`ls $keystore`
        for k in $keys
        do
            keyinfo=`cat ${keystore}/${k} | sed s/[[:space:]]//g`
            keyinfo=${keyinfo,,}sss
            ACCOUNT=${keyinfo:12:40}
            echo "account: ${ACCOUNT}"
            break
        done
    else
        echo "[WARN]: !!! The node has no account !!!"
    fi
}

function check() {
     if [[ $PUBKEY == "" ]];then
        echo "[ERROR]: ********** PUBKEY is empty *********"
        exit
     fi
}

function replace() {
    if [[ $account != "" ]];then
        ACCOUNT=${account}
    fi
    if [[ $p2p_port != "" ]];then
        P2P_PORT=${p2p_port}
    fi
    if [[ $rpc_port != "" ]];then
        RPC_PORT=${rpc_port}
    fi
    if [[ $ip != "" ]];then
        IP=${ip}
    fi
    if [[ $pubkey != "" ]];then
        PUBKEY=${pubkey}
    fi
}

NODE_DIR=${WORKSPACE_PATH}/data/node-${NODE_ID}
readEnv
replace
check


config="${CONF_PATH}/ctool.json"
node_manager_abi="${CONF_PATH}/contracts/nodeManager.cpp.abi.json"

#cns_manager_abi="${CONF_PATH}/contracts/cnsManager.cpp.abi.json"
#cns_addr="0x0000000000000000000000000000000000000011"
#func="getContractAddress"
#param1="__sys_NodeManager"
#param2="latest"
#ret=`${BIN_PATH}/ctool invoke --config $config --abi $cns_manager_abi --addr $cns_addr --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g`
#
#addr=${ret#*result:}
#echo "[INFO]: nodeManager's address: ${addr}"
node_manager_addr="0x1000000000000000000000000000000000000002"

nodeJsonStr_string="{\"name\":\"${NODE_ID}\",\"type\":${NODE_TYPE},\"publicKey\":\"${PUBKEY}\",\"desc\":\"$DESC\",\"externalIP\":\"${IP}\",\"internalIP\":\"${IP}\",\"rpcPort\":${RPC_PORT},\"p2pPort\":${P2P_PORT},\"owner\":\"0x${ACCOUNT}\",\"status\":1}"

${BIN_PATH}/ctool invoke --config $config --addr $node_manager_addr --abi $node_manager_abi --func "add" --param $nodeJsonStr_string

sleep 2

${BIN_PATH}/ctool invoke --config $config --addr $node_manager_addr --abi $node_manager_abi --func getAllNodes
