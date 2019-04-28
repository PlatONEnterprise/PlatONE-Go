#! /bin/bash

address=''

emptyCode='{"jsonrpc":"2.0","id":1,"result":"0x"}'


function deploy() {
    config="./data/config.json"
    code="./data/demo.wasm"
    abi="./data/demo.cpp.abi.json"
    ret=`ctool deploy --config $config --code $code --abi $abi | sed s/[[:space:]]//g`

    address=${ret#*contractaddress:}

#    address="0x7cf06df7bcb5291007ff04f69c179e07a2e1b641"
    if [ ${#address} -ne 42 ]; then
        echo "address is invalid and $name deployed failed."
        exit 1
    fi

    data="{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$address\", \"latest\"],\"id\":1}"
    contractCode=`curl -X POST -H "Content-Type:application/json" --data "$data" http://localhost:6789`

    if [ "$contractCode" == "$emptyCode" ]; then
        echo "$name deployed failed."
    else
        echo "$name deployed succ. Address: $address"
    fi

}

function main() {

    while true
    do
        deploy
        sleep 1h
    done
}

main