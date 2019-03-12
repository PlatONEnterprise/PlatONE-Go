#!/bin/bash

function get_address() {
    name=$1
    config="../conf/ctool.json"
    abi="../conf/contracts/cnsManager.cpp.abi.json"
    addr="0x0000000000000000000000000000000000000011"
    func="getContractAddress"
    param1="__sys_$name"
    param2="latest"
    
    ret=`./ctool invoke --config $config --abi $abi --addr $addr --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g`

    address=${ret#*result:}
    if [ ${#address} -eq 42 ]; then
        echo "$name address: $address"
    else
        echo "$name get address failed."
    fi
}

function main() {
    get_address ParamManager
    get_address UserManager
    get_address UserRegister 
    get_address RoleManager
    get_address RoleRegister
    get_address NodeManager
    get_address NodeRegister
}

main
