#!/bin/bash
cfg='cfg'
config=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
abi=`cat $cfg | grep ABI_JSON | awk -F'=' '{print $2}'`
addr=`cat $cfg | grep ADDR | awk -F'=' '{print $2}'`
ctool=`cat $cfg | grep CTOOL_BIN | awk -F'=' '{print $2}'`

userAddr_string=$1
updateJson_string=$2

$ctool invoke --config $config --addr $addr --abi $abi --func update --param $userAddr_string --param $updateJson_string 
sleep 3
$ctool invoke --config $config --addr $addr --abi $abi --func getAccountByAddress --param $userAddr_string
echo "update"
echo userAddr_string = $userAddr_string
echo updateJson_string = $updateJson_string
