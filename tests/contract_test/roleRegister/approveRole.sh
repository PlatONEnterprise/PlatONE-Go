#!/bin/bash
cfg='cfg'
config=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
abi=`cat $cfg | grep ABI_JSON | awk -F'=' '{print $2}'`
addr=`cat $cfg | grep ADDR | awk -F'=' '{print $2}'`
ctool=`cat $cfg | grep CTOOL_BIN | awk -F'=' '{print $2}'`

address_string=$1
status_int32=$2

$ctool invoke --config $config --addr $addr --abi $abi --func approveRole --param $address_string --param $status_int32 
echo "approveRole"
echo address_string = $address_string
echo status_int32 = $status_int32
