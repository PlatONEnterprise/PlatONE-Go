#!/bin/bash
function help() {
    echo 
    echo "
USAGE: platonectl.sh start [options]
    OPTIONS:
        --nodeid, -n                 start the specified node
        --bootnodes, -b              Connect to the specified bootnodes node
                                     The default is the firstValidatorNode
                                     in genesis.json
        --logsize, -s                Log block size (default: 67108864)
        --logdir, -d                 log dir (default: ../data/node_dir/logs/)
        --all, -a                    start all node
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
RPC_PORT=6791
P2P_PORT=16791
WS_PORT=26791
BOOTNODES=""
EXTRA_OPTIONS=" --debug "
LOG_SIZE=67108864

CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/..
WORKSPACE_PATH=`pwd`
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts
DATA_PATH=${WORKSPACE_PATH}/data

function shiftOption2() {
    if [[ $1 -lt 2 ]];then
        echo "[ERROR]: ********* miss option value! please set the value **********"
        exit
    fi
}

function readConf() {
    conf=${DATA_PATH}/node-${1}/node.conf
    if [[ ! -f $conf ]];then
        return
    fi
    res=`cat $conf | grep "${2}=" | sed -e "s/${2}=\(.*\)/\1/g"`
    if [[ $res == "" ]];then
        return
    fi
    case $2 in
    bootnodes)
       BOOTNODES=$res
       ;;
    logsize)
        LOG_SIZE=$res
    ;;
    logdir)
        LOG_DIR=$res
    ;;
    extraoptions)
        EXTRA_OPTIONS=$res
    ;;
    esac
}

while [ ! $# -eq 0 ]
do
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            ;;
        *)
            help
            exit
            ;;
    esac
    shiftOption2 $#
    shift 2
done

if [[ $NODE_ID == "" ]];then
    exit
fi

NODE_DIR=${WORKSPACE_PATH}/data/node-${NODE_ID}
LOG_DIR=${NODE_DIR}/logs



function readFile() {
    if [ -d ${NODE_DIR} ]; then
        echo "datadir: ${NODE_DIR}"
    else
        echo "[ERROR]: Node does not exist; at node_id=${NODE_ID}"
        exit
    fi

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

    if [ -f ${NODE_DIR}/node.ws_port ]; then
        WS_PORT=`cat ${NODE_DIR}/node.ws_port`
        echo "node.ws_port: ${WS_PORT}"
    else
        echo 'The node.ws_port have not been created'
        exit
    fi

    readConf $NODE_ID "bootnodes"
    if [[ $BOOTNODES == "" ]];then
        if [[ -f ${CONF_PATH}/genesis.json ]];then
            BOOTNODES=`cat ${CONF_PATH}/genesis.json |sed -n '9p'| sed 's/^.*"\(firstValidatorNode\)": "\(.*\)"/\2/g'`
        else
            echo "[ERROR] ************** miss the genesis.json ***************"
        fi
    fi

    readConf $NODE_ID "logsize"
    readConf $NODE_ID "logdir"
    readConf $NODE_ID "extraoptions"


    if [ -d ${LOG_DIR} ]; then
        echo "logdir: ${LOG_DIR}"
    else
        mkdir -p ${LOG_DIR}
        echo "logdir: ${LOG_DIR}"
    fi
}

readFile


echo '
###########################################
####             Start a node          ####
###########################################
'

flag_datadir="--datadir ${NODE_DIR}"
flag_nodekey="--nodekey ${NODE_DIR}/node.prikey"
flag_rpc="--rpc --rpcaddr 0.0.0.0 --rpcport ${RPC_PORT}  --rpcapi db,eth,net,web3,admin,personal,txpool,istanbul "
flag_ws="--ws --wsaddr 0.0.0.0 --wsport ${WS_PORT} "
flag_logs=" --wasmlog  ${LOG_DIR}/wasm_log --wasmlogsize ${LOG_SIZE} "
flag_ipc="--ipcpath ${NODE_DIR}/node-${NODE_ID}.ipc "
flag_pprof=" --pprof --pprofaddr 0.0.0.0 "
flag_gcmode=" --gcmode  archive "

echo "
nohup ${BIN_PATH}/platone --identity platone ${flag_datadir}  --nodiscover \
        --port ${P2P_PORT}  ${flag_nodekey} ${flag_rpc} --rpccorsdomain \""*"\" ${flag_ws} \
        --wsorigins \""*"\" ${flag_logs} ${flag_ipc} \
        --bootnodes ${BOOTNODES} \
        --moduleLogParams '{\"platone_log\": [\"/\"], \"__dir__\": [\"${LOG_DIR}\"], \"__size__\": [\"${LOG_SIZE}\"]}' ${flag_gcmode} ${EXTRA_OPTIONS} \
        1>/dev/null 2>${LOG_DIR}/platone_error.log &
"


ts=`date '+%Y%m%d%H%M%S'`
mkdir -p ${LOG_DIR}
if [ -f ${LOG_DIR}/node-${NODE_ID}.log ]; then
    mv ${LOG_DIR}/node-${NODE_ID}.log ${LOG_DIR}/node-${NODE_ID}.log.bak.$ts
fi

nohup ${BIN_PATH}/platone --identity platone ${flag_datadir}  --nodiscover \
        --port ${P2P_PORT}  ${flag_nodekey} ${flag_rpc} --rpccorsdomain "*" ${flag_ws} \
        --wsorigins "*" ${flag_logs} ${flag_ipc} \
        --bootnodes ${BOOTNODES} \
        --moduleLogParams '{"platone_log": ["/"], "__dir__": ["'${LOG_DIR}'"], "__size__": ["'${LOG_SIZE}'"]}'  ${flag_gcmode}  ${EXTRA_OPTIONS} \
        1>/dev/null 2>${LOG_DIR}/platone_error.log &
sleep 3
