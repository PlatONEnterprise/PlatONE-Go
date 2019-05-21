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
    keyinfo=`./ethkey genkeypair | sed s/[[:space:]]//g`
    keyinfo=${keyinfo,,}
    address=${keyinfo:10:40}
    prikey=${keyinfo:62:64}
    pubkey=${keyinfo:137:128}

    if [ ${#prikey} -ne 64 ]; then
        echo "Error: create node key failed."
        exit
    fi

    datadir=$1
    ts=`date '+%Y%m%d%H%M%S'`
    if [ -f $datadir/node.address ]; then
        mv $datadir/node.address $datadir/node.address.bak.$ts
    fi
    if [ -f $datadir/node.prikey ]; then
        mv $datadir/node.prikey $datadir/node.prikey.bak.$ts
    fi
    if [ -f $datadir/node.pubkey ]; then
        mv $datadir/node.pubkey $datadir/node.pubkey.bak.$ts
    fi

    mkdir -p $datadir
    echo $address > $datadir/node.address
    echo $prikey > $datadir/node.prikey
    echo $pubkey > $datadir/node.pubkey

    echo "Create node key succ. Files: data/node.address, data/node.prikey, data/node.pubkey"
}

function create_account() {
    echo "Input account passphrase."
    ./ethkey generate

    if [ ! -f ./keyfile.json ]; then
        echo "Error: create account failed"
        exit
    fi

    mkdir -p ../../data/keystore
    if [ -f ../../data/keystore/keyfile.json ]; then
        mv ../../data/keystore/keyfile.json ../../data/keystore/keyfile.json.bak.`date '+%Y%m%d%H%M%S'`
    fi
    mv ./keyfile.json ../../data/keystore/keyfile.json

    echo "Create account succ. File: data/keystore/keyfile.json"
}

function create_ctooljson() {
    if [ -f ../conf/ctool.json ]; then
        mv ../conf/ctool.json ../conf/ctool.json.bak.`date '+%Y%m%d%H%M%S'`
    fi
    cp ../conf/ctool.json.template ../conf/ctool.json

    #./repstr ../conf/genesis.json "NODE-KEY" -f ../../data/node.pubkey

    ip=$1
    ./repstr ../conf/ctool.json "NODE-IP" $ip

    keyinfo=`cat ../../data/keystore/keyfile.json | sed s/[[:space:]]//g`
    keyinfo=${keyinfo,,}
    address=${keyinfo:12:40}
    ./repstr ../conf/ctool.json "DEFAULT-ACCOUNT" $address

    echo "Create ctool.json succ. File: conf/ctool.json"
}

function create_genesis() {
    if [ -f ../conf/genesis.json ]; then
        mv ../conf/genesis.json ../conf/genesis.json.bak.`date '+%Y%m%d%H%M%S'`
    fi
    cp ../conf/genesis.json.istanbul.template ../conf/genesis.json

    ./repstr ../conf/genesis.json "NODE-KEY" -f ../../data/node.pubkey

    ip=$1
    ./repstr ../conf/genesis.json "NODE-IP" $ip

    keyinfo=`cat ../../data/keystore/keyfile.json | sed s/[[:space:]]//g`
    keyinfo=${keyinfo,,}
    address=${keyinfo:12:40}
    ./repstr ../conf/genesis.json "DEFAULT-ACCOUNT" $address

    ./ctool codegen --abi ../conf/contracts/cnsManager.cpp.abi.json --code ../conf/contracts/cnsManager.wasm > ../conf/cns-code.hex
    ./repstr ../conf/genesis.json "CNS-CODE" -f ../conf/cns-code.hex
    rm -rf ../conf/cns-code.hex

    echo "Create genesis succ. File: conf/genesis.json"
}

function init_root() {
    if [ -d ../../data/platone ]; then
        echo; echo "Node already initialized, re initailize?"
        #yes_or_no
        echo 
        if [ 2 -ne 1 ]; then
            exit
        fi
    fi

    echo; echo "[Step 1: create node key]"
    if [ -f ../../data/node.pubkey ]; then
        echo "Node key already exists, re create?"
        yes_or_no
        if [ $? -eq 1 ]; then
            create_node_key ../../data
        fi
    else
        create_node_key ../../data
    fi

    echo; echo "[Step 2: create default account]"
    if [ -f ../../data/keystore/keyfile.json ]; then
        echo "Account key file already exists, re create?"
        yes_or_no no
        if [ $? -eq 1 ]; then
            create_account
        fi
    else
        create_account
    fi

    echo; echo "[Step 3: input public ip addr]"
    while true
    do
        #read -p "Your node ip: " ip
        check_ip "127.0.0.1"
        if [ $? -eq 0 ]; then
            break
        else
            echo "Invalid ip. Please re input."
        fi
    done

    echo; echo "[Step 4: create genesis]"
    create_genesis $ip

    echo; echo "[Step 4: create ctool.json]"
    create_ctooljson $ip

    rm -rf ../../data/platone

    echo; echo "[Step 5: init chain data]"
    ./platone --datadir ../../data init ../conf/genesis.json
}

function main() {
    echo "[Nodes initailization]"
    init_root
    rm -rf ../../dataA && cp -rf ../../data ../../dataA && rm ../../dataA/node.*
    rm -rf ../../dataB && cp -rf ../../data ../../dataB && rm ../../dataB/node.*
    rm -rf ../../dataC && cp -rf ../../data ../../dataC && rm ../../dataC/node.*
    create_node_key ../../dataA
    create_node_key ../../dataB
    create_node_key ../../dataC
}

main
