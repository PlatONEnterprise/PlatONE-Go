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

function create_node_cert(){

    echo '
    ++++++++++++++++++++++
    create node cert
    ++++++++++++++++++++++
    '

    ${BIN_PATH}/platonecli ca generateKey --file ${NODE_DIR}/nodekey.pem --curve secp256k1 --target private --format PEM 
    ${BIN_PATH}/platonecli ca generateCSR --organization wxbc --commonName ${NODE_ID} --dgst sha256 --keyfile ${NODE_DIR}/nodekey.pem --file ${NODE_DIR}/node.csr

    ${BIN_PATH}/platonecli  ca create  --ca ${CA_PATH}/org.crt --csr ${NODE_DIR}/node.csr  --keyfile ${CA_PATH}/orgkey.pem --serial 10 --file ${NODE_DIR}/node.crt

    cat ${NODE_DIR}/node.crt
}

function create_node_key() {
    keyinfo=`${BIN_PATH}/ethkey genkeypair | sed s/[[:space:]]//g`
    keyinfo=${keyinfo,,}
    address=${keyinfo:10:40}
    prikey=${keyinfo:62:64}
    pubkey=${keyinfo:137:128}

    if [ ${#prikey} -ne 64 ]; then
        echo "[ERROR]: create node key failed."
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

    echo "[INFO]: Create node key succ. Files: ${NODE_DIR}/node.address, ${NODE_DIR}/node.prikey, ${NODE_DIR}/node.pubkey"
}

function checkGenesis() {
    if [ -f ${CONF_PATH}/genesis.json ]; then
        return 1
    else 
        return 0
    fi    
}


function setup_node_datadir() {
    if [ "${AUTO}" = "true" ]; then 
        echo "[INFO]: auto setup node dir"
        if [ -f ${NODE_DIR}/node.pubkey ]; then
            return
        fi
        create_node_key
        return
    else
        echo "AUTO: ${AUTO}"  
    fi    

    echo; echo "[INFO]: Do You What To Create a new node key ? (Please do not recreate the first node node.key)"
    yes_or_no
    if [ $? -eq 1 ]; then
        if [ -f ${NODE_DIR}/node.pubkey ]; then
            echo "[INFO]: Node key already exists, re create? (Please do not recreate the first node node.key)"
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
            echo "[INFO]: Node's public key: ${pubkey}"
        else
            echo; echo "!!! No node's public key file !!!"
            echo "[WARN]: Please Put Your Nodekey file \"node.pubkey\" to the directory ${NODE_DIR}"
            exit
        fi
    fi
}

function main() {
    checkGenesis
    if [ $? -eq 0 ]; then
        echo '[ERROR]: !!! There No Genesis File, Please Generate Genesis File First !!!'
        exit
    fi

    if [ -d ${NODE_DIR} ]; then
        echo "datadir: ${NODE_DIR}"
    else
        echo '[INFO]: The node directory have not been created, Now to create it'
        mkdir -p ${NODE_DIR}
    fi
    setup_node_datadir

    create_node_cert

    ${BIN_PATH}/platone --datadir ${NODE_DIR} init  ${CONF_PATH}/genesis.json
}

function help() {
    echo 
    echo "
USAGE: platonectl.sh init [options]

        OPTIONS:
            --nodeid, -n                 set node id (default=0)
            --ip                         set node ip (default=127.0.0.1)
            --rpc_port                   set node rpc port (default=6791)
            --p2p_port                   set node p2p port (default=16791)
            --ws_port                    set node ws port (default=26791)
            --auto                       auto=true: will no prompt to create
                                         the node key and init (default: false)
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
WS_PORT=26791
AUTO=false

CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts
CA_PATH=${WORKSPACE_PATH}/ca-certs


while [ ! $# -eq 0 ]
do
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            ;;
        --ip )
            echo "ip: $2"
            IP=$2
            ;;
        --rpc_port)
            echo "rpc_port: $2"
            RPC_PORT=$2
            ;;
        --ws_port)
            echo "ws_port: $2"
            WS_PORT=$2
            ;;
        --p2p_port)
            echo "p2p_port: $2"
            P2P_PORT=$2
            ;;
        --auto)
            echo "auto $2"
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

if [ -d ${NODE_DIR} ]; then
    echo "datadir: ${NODE_DIR}"
else
    echo '[INFO]: The node directory have not been created. Now to create it'
    mkdir -p ${NODE_DIR}
fi

echo ${IP} > ${NODE_DIR}/node.ip
echo ${P2P_PORT} > ${NODE_DIR}/node.p2p_port
echo ${WS_PORT} > ${NODE_DIR}/node.ws_port
echo ${RPC_PORT} > ${NODE_DIR}/node.rpc_port

echo '
###########################################
####             Init a node           ####
###########################################
'

main
