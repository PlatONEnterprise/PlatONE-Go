#!/bin/bash

SCRIPT_NAME="$(basename ${0})"
CURRENT_PATH=`pwd`
WORKSPACE_PATH=`cd ${CURRENT_PATH}/.. && echo ${PWD}`
BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/script
DATA_PATH=${WORKSPACE_PATH}/data
ENABLE=""
DISENABLE=""
NODE_ID=0
cd ${CURRENT_PATH}

VERSION=`${BIN_PATH}/platone --version`

function usage() {
cat <<EOF
#h0    DESCRIPTION
#h0        The deployment script for platone
#h0
#h0    Usage:
#h0        ${SCRIPT_NAME} <command> [command options] [arguments...]
#h0
#h0    COMMANDS
#c0        init                             initialize node. please setup genesis first
#c1        init OPTIONS
#c1            --nodeid, -n                 set node id (default=0)
#c1            --ip                         set node ip (default=127.0.0.1)
#c1            --rpc_port                   set node rpc port (default=6791)
#c1            --p2p_port                   set node p2p port (default=16791)
#c1            --ws_port                    set node ws port (default=26791)
#c1            --auto                       auto=true: will no prompt to create
#c1                                         the node key and init (default: false)
#c1            --help, -h                   show help
#c1
#c1        example: ${SCRIPT_NAME} init -n 1
#c1                                    --ip 127.0.0.1
#c1                                    --rpc_port 6790
#c1                                    --p2p_port 16790
#c1                                    --ws_port 26790
#c1                                    --auto "true"
#c1                 or:
#c1                     ${SCRIPT_NAME} init
#c0        one                              start a node completely
#c0                                         default account password: 0
#c0        four                             start four node completely
#c0                                         default account password: 0
#c0        start                            try to start the specified node
#c4        start OPTIONS
#c4            --nodeid, -n                 start the specified node
#c4            --bootnodes, -b              Connect to the specified bootnodes node
#c4                                         The default is the first in the suggestObserverNodes
#c4                                         in genesis.json
#c4            --logsize, -s                Log block size (default: 67108864)
#c4            --logdir, -d                 log dir (default: ../data/node_dir/logs/)
#c4                                         The path connector '/' needs to be escaped
#c4                                         when set: eg ".\/logs"
#c4            --extraoptions, -e           extra platone command options when platone starts
#c4                                         (default: --debug)
#c4            --all, -a                    start all node
#c4            --help, -h                   show help
#c0        stop                             try to stop the specified node
#c5        stop OPTIONS
#c5            --nodeid, -n                 stop the specified node
#c5            --all, -a                    stop all node
#c5            --help, -h                   show help
#c0        restart                          try to restart the specified node
#c6        restart OPTIONS
#c6            --nodeid, -n                 restart the specified node
#c6            --all, -a                    restart all node
#c6            --help, -h                   show help
#c0        console                          start an interactive JavaScript environment
#c7        console OPTIONS
#c7            --opennodeid , -n            open the specified node console
#c7                                         set the node id here
#c7            --closenodeid, -c            stop the specified node console
#c7                                         set the node id here
#c7            --closeall                   stop all node console
#c7            --help, -h                   show help
#c0        deploysys                        deploy the system contract
#c8        deploysys OPTIONS
#c8            --nodeid, -n                 the specified node id (default: 0)
#c8            --auto                       auto=true: will use the default node password: 0
#c8                                         to create the account and also
#c8                                         to unlock the account (default: false)
#c8            --help, -h                   show help
#c0        updatesys                        normal node update to consensus node
#c9        updatesys OPTIONS
#c9            --nodeid, -n                 the specified node id
#c9            --content, -c                update content (default: '{"type":1}')
#c9            --help, -h                   show help
#c0        addnode                          add normal node to system contract
#c10       addnode OPTIONS
#c10           --nodeid, -n                 the specified node id. must be specified
#c10           --desc                       the specified node desc
#c10           --p2p_port                   the specified node p2p_port
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --rpc_port                   the specified node rpc_port
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --ip                         the specified node ip
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --pubkey                     the specified node pubkey
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --account                    the specified node account
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --help, -h                   show help
#c0        clear                            clear all nodes data
#c11       clear OPTIONS
#c11           --nodeid, -n                 clear specified node data
#c11           --all, -a                    clear all nodes data
#c11           --help, -h                   show help
#c0        unlock                           unlock node account
#c12       unlock OPTIONS
#c12           --nodeid, -n                 unlock account on specified node
#c12           --account, -a                account to unlock
#c12           --phrase, -p                 phrase of the account
#c12           --help, -h                   show help
#c0        get                              display all nodes in the system contract
#c0        setupgen                         create the genesis.json and compile sys contract
#c13       setupgen OPTIONS
#c13           --nodeid, -n                 the first node id (default: 0)
#c13           --ip                         the first node ip (default: 127.0.0.1)
#c13           --p2p_port                   the first node p2p_port (default: 16791)
#c13           --auto                       auto=true: Will auto create new node keys and will
#c13                                        not compile system contracts again (default=false)
#c13           --observerNodes, -o          set the genesis suggestObserverNodes
#c13                                        (default is the first node enode code)
#c13           --validatorNodes, -v         set the genesis validatorNodes
#c13                                        (default is the first node enode code)
#c13           --interpreter, -i            Select virtual machine interpreter in wasm, evm, all (default: wasm)
#c13           --help, -h                   show help
#c0        status                           show all node status
#c14       status OPTIONS                   show all node status
#c14           --nodeid, -n                 show the specified node status info
#c14           --all, -a                    show all  node status info
#c14           --help, -h                   show help
#c0        createacc                        create account
#c15       createacc OPTIONS
#c15           --nodeid, -n                 create account for specified node
#c15           --help, -h                   show help
#c0        version                          show platone release version
#c0===============================================================
#c0    INFORMATION
#c0        version         PlatONE Version: ${VERSION}
#c0        author
#c0        copyright       Copyright
#c0        license
#c0
#c0===============================================================
#c0    HISTORY
#c0        2019/06/26  ty : create the deployment script
#c0
#c0===============================================================
EOF
}

function showUsage() {
    if [[ $1 == "" ]];then
        usage |  grep -e "^#[ch]0 " | sed -e "s/^#[ch][0-9]*//g" ;
        return
    fi
    usage |  grep -e "^#h0 \|^#c${1} " | sed -e "s/^#[ch][0-9]*//g" ;
}

function shiftOption2() {
    if [[ $1 -lt 2 ]];then
        echo "[ERROR]: ********* miss option value! please set the value **********"
        exit
    fi
}

function helpOption() {
    for op in "$@"
    do
        if [[ $op == "--help" ]] || [[ $op == "-h" ]]; then
            return 1
        fi
    done
}

function init() {
    helpOption "$@"
    if [[ $? -ne 0 ]];then
        showUsage 1
        return
    fi
    ./init-node.sh "$@"
}

function one() {
    ./setup-genesis.sh --auto "true"      # 不加auto 会重新编译合约
    ./init-node.sh --auto "true"          # 不加auto 将会创建nodekey的时候进行提示
    ./start-node.sh
    ./deploy-system-contract.sh  --auto "true"  # 不加auto 将不会使用默认的密码0 创建和解锁账号; 此时会创建第一个账号; 部署会默认把第一节点加入合约
}

function four() {
    ./build-4-nodes-chain.sh "$@"
}

function nodeIsRunning() {
     if [[ `ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${1} " | grep -v grep | awk '{print $2}'` != "" ]]; then
        return 1
     fi
     return 0
}

function checkNodeStatusFullName() {
     if [[ -d ${DATA_PATH}/${1} ]] && [[ $1 == node-* ]]; then
        nodeid=`echo ${1#*-}`
        if [[ `ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${nodeid} " | grep -v grep | awk '{print $2}'` != "" ]]; then
            ENABLE=`echo "${ENABLE} ${nodeid}"`
        else
            DISENABLE=`echo ${DISENABLE} ${nodeid}`
        fi
    fi
}

function checkAllNodeStatus() {
    nodes=`ls ${DATA_PATH}`
    for n in ${nodes}
    do
        checkNodeStatusFullName $n
    done
}


function saveConf() {
    node_conf=${DATA_PATH}/node-${1}/node.conf
    node_conf_tmp=${DATA_PATH}/node-${1}/node.conf1
    if [[ $3 == "" ]];then
        return
    fi

    if ! [[ -f "${node_conf}" ]];then
        touch "${node_conf}"
    fi

    res=`cat ${node_conf} | grep "$2"`
    if [[ ${res} != "" ]];then
        cat $node_conf | sed "s#${2}=.*#${2}=${3}#g" | cat > $node_conf_tmp
        mv $node_conf_tmp $node_conf
    else
        echo "${2}=${3}" >> ${node_conf}
    fi
}

function start() {
    nid=""
    bns=""
    logsize=""
    logdir=""
    extraoptions=""
    all="false"
    if [[ $# -eq 0 ]];then
         showUsage 4
         exit
    fi
    while [ ! $# -eq 0 ]
    do
        case "$1" in
        --nodeid | -n)
            shiftOption2 $#
            nodeIsRunning $2
            if [[ $? -ne 0 ]]; then
                echo "[WARN]: the node is running ...; node_id: $2"
                return
            fi
            echo "[INFO]: start node: ${2}"
            nid=$2
            shift 2
            ;;
        --bootnodes | -b)
            shiftOption2 $#
            bns=$2
            shift 2
            ;;
        --logsize | -s)
            shiftOption2 $#
            logsize=$2
            shift 2
            ;;
        --logdir | -d)
            shiftOption2 $#
            logdir=$2
            shift 2
            ;;
        --extraoptions | -e)
            shiftOption2 $#
            extraoptions=$2
            shift 2
            ;;
        --all | -a)
            echo "[INFO]: start all nodes"
             all=true
             shift 1
            ;;
        *) showUsage 4; exit;;
        esac
    done

    if [[ $all == true ]];then
        checkAllNodeStatus
        for d in ${DISENABLE};do
            echo "[INFO]: start all disable nodes"
            saveConf $d bootnodes "${bns}"
            saveConf $d logsize "${logsize}"
            saveConf $d logdir "${logdir}"
            saveConf $d extraoptions "${extraoptions}"
            ./start-node.sh -n $d
        done
        exit
    fi
    saveConf $nid bootnodes "${bns}"
    saveConf $nid logsize "${logsize}"
    saveConf $nid logdir "${logdir}"
    saveConf $nid extraoptions "${extraoptions}"
    ./start-node.sh -n $nid
}

function stop() {
    stopAll(){
        nodes=`ls ${DATA_PATH}`
        for n in ${nodes}
        do
            if [[ -d ${DATA_PATH}/$n ]] && [[ $n == node-* ]]; then
                nodeid=`echo ${n#*-}`
                stop --nodeid $nodeid
            fi
        done
        killall "platone"
    }

    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        pid=`ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${2} " | grep -v grep | awk '{print $2}'`
        if [[ $pid != "" ]]; then
            echo "[INFO]: stop node: ${2}"
            kill $pid
            sleep 1
        fi
        ;;
    --all | -a)
        echo "[INFO]: stop all nodes"
        stopAll
        ;;
    *) showUsage 5;;
    esac
}

function console() {
    case "$1" in
    --opennodeid | -n)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=`cat node.rpc_port`
        ip=`cat node.ip`
        cd ${BIN_PATH}
        ./platone attach http://${ip}:${rpc_port}
        cd ${CURRENT_PATH}
        ;;
    --closenodeid | -c)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=`cat node.rpc_port`
        ip=`cat node.ip`
        pid=`ps -ef | grep "platone attach http://${ip}:${rpc_port}" | grep -v grep | awk '{print $2}'`
        cd ${CURRENT_PATH}
        ;;
    --closeall)
        killall "platone attach"
        ;;
    *) showUsage 7;;
    esac
}

function deploySys() {
    helpOption "$@"
    if [[ $? -ne 0 ]];then
        showUsage 8
        return
    fi
    ./deploy-system-contract.sh "$@"
}

function updateSys() {
    helpOption "$@"
    if [[ $? -ne 0 ]];then
        showUsage 9
        return
    fi
    ./update_to_consensus_node.sh "$@"
}

function addNode() {
    helpOption "$@"
    if [[ $? -ne 0 ]];then
        showUsage 10
        return
    fi
    ./add-node.sh "$@"
}

function getInformation() {
    if [[ -f ${1}/${2} ]]; then
        echo $(cat ${1}/${2})
    fi
}

function echoInformation() {
    res=`getInformation $1 $2`
    echo "                  ${2}: ${res}"
}

function showNodeInformation() {
    nodeHome=${DATA_PATH}/node-${1}
    echo "          node info:"

    keystore=${nodeHome}/keystore
    if [[ -d $keystore ]];then
        keys=`ls $keystore`
        for k in $keys
        do
            keyinfo=`cat ${keystore}/${k} | sed s/[[:space:]]//g`
            keyinfo=${keyinfo,,}sss
            account=${keyinfo:12:40}
            echo "                  account: ${account}"
            break
        done
    fi

    echoInformation ${nodeHome} node.address
    echoInformation ${nodeHome} node.ip
    echoInformation ${nodeHome} node.rpc_port
    echoInformation ${nodeHome} node.p2p_port
    echoInformation ${nodeHome} node.ws_port
    echoInformation ${nodeHome} node.pubkey
}

function showNodeInformation1() {
    case $1 in
    enable)
        echo "running -> node_id:  ${2}"
       ;;
    disable)
        echo "disable -> node_id:  ${2}"
       ;;
    esac
    showNodeInformation $2
}

function show() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-$2"
        ;;
    --all | -a)
        checkAllNodeStatus
        ;;
    *) showUsage 14;;
    esac
    for e in ${ENABLE} ; do
        showNodeInformation1 enable $e
    done

    for d in ${DISENABLE}; do
        showNodeInformation1 disable $d
    done
}

function restart() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-${2}"
        if [[ ${ENABLE} == "" ]];then
            echo "[WARN]: the node is not running"
            echo "[INFO]: to start the node..."
            start -n $2
            exit
        fi
        stop -n $2
        start -n $2
        ;;
    --all | -a)
        echo "[INFO]: restart all running nodes"
        checkAllNodeStatus
        for e in ${ENABLE}; do
            stop -n $e
            start -n $e
        done

        ;;
    *) showUsage 6;;
    esac
}

function getAllNodes() {
    config="${CONF_PATH}/ctool.json"
    abi="${CONF_PATH}/contracts/nodeManager.cpp.abi.json"
#    abiCNS="${CONF_PATH}/contracts/cnsManager.cpp.abi.json"
#    addrCNS="0x0000000000000000000000000000000000000011"
#    func="getContractAddress"
#    param1="__sys_NodeManager"
#    param2="latest"
#
#    echo "${BIN_PATH}/ctool invoke --config $config --abi $abiCNS --addr $addrCNS --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g"
#    ret=`${BIN_PATH}/ctool invoke --config $config --abi $abiCNS --addr $addrCNS --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g`
#    addr=${ret#*result:}

    node_manager_addr="0x1000000000000000000000000000000000000002"
    echo "[INFO]: get all nodes from nodeManager system contract"
    ${BIN_PATH}/ctool invoke --config $config --addr $node_manager_addr --abi $abi --func getAllNodes
}

function clearConf() {
    if [[ -f ${CONF_PATH}/${1} ]]; then
        mkdir -p ${CONF_PATH}/bak
        mv ${CONF_PATH}/${1} ${CONF_PATH}/bak/${1}.bak.`date '+%Y%m%d%H%M%S'`
    fi
}

function clear() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        stop -n $2
        echo "[INFO]: clear node id: ${2}"
        NODE_DIR=${WORKSPACE_PATH}/data/node-${2}
        echo "[INFO]: clean NODE_DIR: ${NODE_DIR}"
        rm -rf ${NODE_DIR}
        ;;
    --all | -a)
        stop -a
        echo "[INFO]: clear all nodes data"
        data=${WORKSPACE_PATH}/data
        rm -rf ${data}/*
        clearConf ctool.json
        clearConf genesis.json
        ;;
    *) showUsage 11;;
    esac
}

function unlockAccount() {
    IP=${1}
    PORT=${2}
    account=${3}
    pw=${4}
    
    echo "[INFO]: unlock node account, nodeid: ${NODE_ID}"

    if [ -z ${account} ]
    then 
        # get node owner address
        keystore=${DATA_PATH}/node-${NODE_ID}/keystore/
        echo $keystore
        keys=`ls $keystore`
        echo "$keys"
        for k in $keys
        do
            keyinfo=`cat $keystore/$k | sed s/[[:space:]]//g`
            keyinfo=${keyinfo,,}sss
            account=${keyinfo:12:40}
            account="0x${account}"
            echo "account: ${account}"
            break
        done
    fi

    if [ -z ${pw} ]
    then
        read -p "Your account password?: " pw
    fi

    echo "curl -X POST  -H 'Content-Type: application/json' --data '{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"${account}\",\"${pw}\",0],\"id\":1}' http://${1}:${2}"
    curl -X POST  -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"${account}\",\"${pw}\",0],\"id\":1}" http://${1}:${2}
}

function unlock() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        nodeHome=${DATA_PATH}/node-${2}
        NODE_ID=$2
        echo $nodeHome
        ip=`getInformation ${nodeHome} node.ip`
        rpc_port=`getInformation ${nodeHome} node.rpc_port`
        echo ${ip} ${rpc_port}
        unlockAccount ${ip} ${rpc_port}
        ;;
    *) showUsage 12; exit;;
    esac
}


function unlockAcc(){

    ACC=""
    PHRASE=""

    while [ ! $# -eq 0 ]
    do
        case "$1" in
            --nodeid | -n)
                nodeHome=${DATA_PATH}/node-${2}
                NODE_ID=$2
                echo $nodeHome
                ip=`getInformation ${nodeHome} node.ip`
                rpc_port=`getInformation ${nodeHome} node.rpc_port`
                echo ${ip} ${rpc_port}
                ;;
            --account | -a)
                echo "account: $2"
                ACC=${2}
                ;;
            --phrase | -p)
                PHRASE=${2}
                ;;
            *)
                showUsage 12
                exit
                ;;
        esac
        shiftOption2 $#
        shift 2
    done

    unlockAccount ${ip} ${rpc_port} ${ACC} ${PHRASE}
}

function setupGenesis() {
    helpOption "$@"
    if [[ $? -ne 0 ]];then
        showUsage 13
        return
    fi
    ./setup-genesis.sh "$@"
}

function create_account() {
    echo "Input account passphrase."
    read -p "passphrase: " pw
    echo "$1:$2"
    ret=$( curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${pw}\"],\"id\":1}"  http://${1}:${2} )
    echo ${ret}

    substr=${ret##*\"result\":\"}
    if [[ ${#substr} -gt 42 ]]; then
        ACCOUNT=${substr:0:42}
        echo "New account: "${ACCOUNT}
    else
        echo "[ERROR]: create account failed!!! Check if node has started"
        exit
    fi
}

function createAcc() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        nodeHome=${DATA_PATH}/node-${2}
        echo $nodeHome
        ip=`getInformation ${nodeHome} node.ip`
        rpc_port=`getInformation ${nodeHome} node.rpc_port`
        echo ${ip} ${rpc_port}
        create_account ${ip} ${rpc_port}
        ;;
    *) showUsage 15;;
    esac
}

function showVersion() {
    echo "${VERSION}"
}

case $1 in
init) shift; init "$@";;
one) shift; one;;
four) shift; four "$@";;
stop) shift; stop "$@";;
start) shift; start "$@";;
restart) shift; restart "$@";;
console) shift; console "$@";;
deploysys) shift; deploySys "$@";;
updatesys) shift; updateSys "$@";;
createacc) shift; createAcc "$@";;
setupgen) shift; setupGenesis "$@";;
addnode) shift; addNode "$@";;
unlock) shift; unlockAcc "$@";;
version | -v) showVersion;;
status) shift; show "$@";;
clear) shift; clear "$@";;
get) shift; getAllNodes;;
*) shift; showUsage;;
esac


