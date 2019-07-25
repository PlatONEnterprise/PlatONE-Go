#!/bin/bash
function unlock_account() {
    http_data="{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"$1\",\"$2\",0],\"id\":1}"
    #echo $http_data
    curl -H "Content-Type: application/json" --data "${http_data}"  http://${IP}:${RPC_PORT}
}

function make_config_json() {
    echo "{
  \"url\":\"http://NODE-IP:RPC-PORT\",
  \"gas\":\"0x0\",
  \"gasPrice\":\"0x0\",
  \"from\":\"0xDEFAULT-ACCOUNT\"
}

" > ${CURRENT_PATH}/config.json

    ${BIN_PATH}/repstr ${CURRENT_PATH}/config.json "NODE-IP" ${IP}
    ${BIN_PATH}/repstr ${CURRENT_PATH}/config.json "RPC-PORT" ${RPC_PORT}
    echo "${IP}:${RPC_PORT}"

    ${BIN_PATH}/repstr ${CURRENT_PATH}/config.json "DEFAULT-ACCOUNT" ${ACCOUNT:2}
    

    echo;echo " Create config.json for contract-deploy"
    echo ""

    cat ${CURRENT_PATH}/config.json
}


function create_account() {
    phrase=0

    if [ "${AUTO}" = "true" ]; then 
        echo "auto build chain" 
    else
        echo "Input account passphrase."
        read -p "passphrase: " phrase
    fi
    #echo "$IP:$RPC_PORT"
    ret=$( curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${phrase}\"],\"id\":1}"  http://${IP}:${RPC_PORT} )
    echo $ret

    substr=${ret##*\"result\":\"}

    if [ ${#substr} -gt 42 ]; then
        ACCOUNT=${substr:0:42}
        echo "New account: "${ACCOUNT}
        #echo "passphrase: "${phrase}
        unlock_account ${ACCOUNT} ${phrase}
        make_config_json ${ACCOUNT} $IP $RPC_PORT
    else
        echo "create account failed!!!"
        exit
    fi
}


function help() {
    echo 
    echo "
${SCRIPT_NAME} [options]

For Example: 
            ${SCRIPT_NAME} --ip 127.0.0.1 --rpc_port 6791
    --ip        ip address of the node to connect, default 127.0.0.1
    --rpc_port  rpc port of the node to connect, default 6791
"
}

# variables
SCRIPT_NAME=$0
SCRIPT_DIR_R=`dirname "${SCRIPT_NAME}"`
CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR_R}
SCRIPT_DIR=`pwd`
cd ${CURRENT_PATH}

#SCRIPT_DIR="$( cd -P "$( dirname "${SCRIPT_NAME}" )" >/dev/null 2>&1 && pwd )"
IP=127.0.0.1
RPC_PORT=6791
AUTO=false

ACCOUNT=""

CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/../..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/chain/PlatONE_linux/bin

while [ ! $# -eq 0 ]
do
    case "$1" in
        --help | -h)
            help
            exit
            ;;
        --ip )
            echo "nodeid: $2"
            IP=$2
            #exit
            ;;                
        --rpc_port)
            echo "rpc_port: $2"
            RPC_PORT=$2
            #exit
            ;;
        --auto)
            AUTO=$2
            ;;
        --bootnodes)
            echo "bootnodes: $2"
            BOOTNODES=$2
            #exit
            ;;
    esac
    shift
done

echo '
###########################################
####       Create an account           ####
###########################################
'

create_account
