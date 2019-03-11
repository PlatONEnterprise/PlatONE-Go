#!/bin/bash

function deploy() {
    name=$1
    config="../conf/ctool.json"
    code="../conf/contracts/$name.wasm"
    abi="../conf/contracts/$name.cpp.abi.json"

    ret=`./ctool deploy --config $config --code $code --abi $abi | sed s/[[:space:]]//g`

    address=${ret#*contractaddress:}
    if [ ${#address} -eq 42 ]; then
        echo "$name deployed succ. Address: $address"
    else
        echo "$name deployed failed."
    fi
}

function main() {
    deploy paramManager
    deploy userManager 
    deploy userRegister 
    deploy roleManager
    deploy roleRegister 
    deploy nodeManager 
    deploy nodeRegister 
}

main
