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

function create_ca_cert() {
    mkdir -p ${WORKSPACE_PATH}/ca-certs

    ${BIN_PATH}/platonecli ca generateKey --file ${CA_PATH}/rootkey.pem --curve secp256k1 --target private --format PEM
    ${BIN_PATH}/platonecli ca genSelfSignCert --organization wxbc --commonName rootCA --dgst sha256 --serial 1 --file ${CA_PATH}/root.crt --private ${CA_PATH}/rootkey.pem

    ${BIN_PATH}/platonecli ca generateKey --file ${CA_PATH}/orgkey.pem --curve secp256k1 --target private --format PEM 
    ${BIN_PATH}/platonecli ca generateCSR --organization wxbc --commonName defaultOrg --dgst sha256 --private ${CA_PATH}/orgkey.pem --file ${CA_PATH}/org.csr
    ${BIN_PATH}/platonecli  ca create  --ca  ${CA_PATH}/root.crt --csr ${CA_PATH}/org.csr  --private ${CA_PATH}/rootkey.pem --serial 100 --file ${CA_PATH}/org.crt
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
         replaceList "__VALIDATORS__" $VALIDATOR_NODES
    else
         replaceList "__VALIDATORS__" $default_enode
    fi
    if [[ $OBSERVE_NODES != "" ]]; then
         replaceList "__OBSERVES__" $OBSERVE_NODES
    else
         replaceList "__OBSERVES__" $default_enode
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

function generateKey(){
    if [ "$OUT_FILE" = "" ]; then 
        echo "must input the out file name!"
        return
    fi

    outdir=`(dirname "${OUT_FILE}")`
    if [ ! -d "$outdir" ]; then
        mkdir -p $outdir
    fi

    ${BIN_PATH}/platonecli ca generateKey --file ${OUT_FILE} --curve secp256k1 --target private --format PEM
}

function genSelfSignCert(){
    if [ "$PRIVATE" = "" ]; then
        echo "must input private key file!"
        return 
    fi 

    if [ ! -f $PRIVATE ]; then 
        echo "private key file not exist!"
        return
    fi

    if [ "$OUT_FILE" = "" ]; then 
        echo "must input the out file name!"
        return
    fi

    outdir=`(dirname "${OUT_FILE}")`
    if [ ! -d "$outdir" ]; then
        mkdir -p $outdir
    fi

    ${BIN_PATH}/platonecli ca genSelfSignCert --organization $ORG --commonName $CNAME --dgst sha256 --serial 1 --file $OUT_FILE --private $PRIVATE
}

function genRequest() {
    if [ "$PRIVATE" = "" ]; then
        echo "must input private key file!"
        return 
    fi 

    if [ ! -f $PRIVATE ]; then 
        echo "private key file not exist!"
        return
    fi

    if [ "$OUT_FILE" = "" ]; then 
        echo "must input the out file name!"
        return
    fi

    outdir=`(dirname "${OUT_FILE}")`
    if [ ! -d "$outdir" ]; then
        mkdir -p $outdir
    fi

    ${BIN_PATH}/platonecli ca generateCSR --organization $ORG --commonName $CNAME --dgst sha256  --file $OUT_FILE --private $PRIVATE
}

function createCert() {
    if [ "$PRIVATE" = "" ]; then
        echo "must input private key file!"
        return 
    fi 

    if [ ! -f $PRIVATE ]; then 
        echo "private key file not exist!"
        return
    fi

    if [ "$OUT_FILE" = "" ]; then 
        echo "must input the out file name!"
        return
    fi

    if [ "$CSR_FILE" = "" ]; then 
        echo "must input the csr file name!"
        return
    fi


    outdir=`(dirname "${OUT_FILE}")`
    if [ ! -d "$outdir" ]; then
        mkdir -p $outdir
    fi

    ${BIN_PATH}/platonecli  ca create  --ca  $CA --csr $CSR_FILE  --private $PRIVATE --serial 100 --file $OUT_FILE
}


###########################################
#### Ca-cert operations ####
###########################################

function help() {
    echo 
    echo "
USAGE: platonectl.sh setupgen [options]

        OPTIONS:
           --genkey                     generate key pair
           --selfsign                   private key of issuer
           --request                    generate cert request file
           --create                     create cert for request
           --private                    private key of issuer
           --ca                         cert of isuuer
           --out                        output of cert file
           --type                       type of cert(root, org or node)
           --nodeid, -n                 the first node id (default: 0)
           --org                        orgnization of the cert
           --cname                      common name of the cert
           --auto                       auto=true: Will auto create new node keys and will
                                        not compile system contracts again (default=false)
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
CA_PATH=${WORKSPACE_PATH}/ca-certs

GENKEY=false
SELFSIGN=false
REQUEST=false 
CREATE=false
PRIVATE=""
CA=""
OUT_FILE=""
CSR_FILE=""
CERT_TYPE=""
ORG="PlatONE-ORG"
CNAME="PlatONE-CNAME"

while [ ! $# -eq 0 ]
do
    shiftCnt=1
    case "$1" in
        --nodeid | -n)
            echo "nodeid: $2"
            NODE_ID=$2
            shiftCnt=2
            ;;
        --genkey | -g)
            echo "genkey: true"
            GENKEY=true
            shiftCnt=1
            ;;
        --selfsign)
            echo "selfsign: true"
            SELFSIGN=true
            shiftCnt=1
            ;;
        --request)
            echo "request: true"
            REQUEST=true
            shiftCnt=1
            ;;        
        --create)
            echo "request: true"
            CREATE=true
            shiftCnt=1
            ;;  
        --private | -p)
            echo "private: $2"
            PRIVATE=$2
            shiftCnt=2
            ;;
        --csr)
            echo "csr: $2"
            CSR_FILE=$2
            shiftCnt=2
            ;;        
        --ca)
            echo "ca: $2"
            CA=$2
            shiftCnt=2
            ;;
        --type)
            echo "cert type: $2"
            CERT_TYPE=$2
            shiftCnt=2
            ;;            
        --org)
            echo "orgnization: $2"
            ORG=$2
            shiftCnt=2
            ;; 
        --cname)
            echo "common name: $2"
            CNAME=$2
            shiftCnt=2
            ;;             
        --out)
            echo "output: $2"
            OUT_FILE=$2
            shiftCnt=2
            ;;         
        --auto)
            echo "auto: $2"
            AUTO=$2
            shiftCnt=2
            ;;
        *)
            help
            exit
            ;;
    esac
    # shiftOption2 $#
    shift $shiftCnt
done

# NODE_DIR=${WORKSPACE_PATH}/data/node-${NODE_ID}

# if [ -d ${NODE_DIR} ]; then
#     echo "root node datadir: ${NODE_DIR}"
# else
#     echo '[INFO]: The node directory have not been created, Now to create it'
#     mkdir -p ${NODE_DIR}
# fi


if [ "$GENKEY" = true ] ; then
    generateKey
    exit
fi

if [ "$SELFSIGN" = true ]; then
    genSelfSignCert
    exit
fi 

if [ "$REQUEST" = true ]; then
    genRequest
    exit
fi 

if [ "$CREATE" = true ]; then
    createCert
    exit
fi 

