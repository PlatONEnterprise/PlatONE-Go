#! /bin/bash

address=''

function deploy() {
    config="./data/config.json"
    code="./data/demo.wasm"
    abi="./data/demo.cpp.abi.json"
    ret=`ctool deploy --config $config --code $code --abi $abi | sed s/[[:space:]]//g`

    address=${ret#*contractaddress:}
    if [ ${#address} -eq 42 ]; then
        echo "$name deployed succ. Address: $address"
    else
        echo "$name deployed failed."
    fi
}

function main() {
    deploy
    ./perf_test -useWs -stressTest=1 -abiPath="./data/demo.cpp.abi.json" -configPath="./data/config.json"  -contractAddress="$address" -totalCount=200000 -realtimeTps=true -consensusTest=true -chanValue=20 -blockDuration=10000 
}

main