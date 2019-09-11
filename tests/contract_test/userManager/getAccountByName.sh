#!/bin/bash
cfg='cfg'
config=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
abi=`cat $cfg | grep ABI_JSON | awk -F'=' '{print $2}'`
addr=`cat $cfg | grep ADDR | awk -F'=' '{print $2}'`
ctool=`cat $cfg | grep CTOOL_BIN | awk -F'=' '{print $2}'`

userJsonStr_string=$1

$ctool invoke --config $config --addr $addr --abi $abi --func getAccountByName --param $userJsonStr_string 
echo "getAccountByName"
echo userJsonStr_string = $userJsonStr_string
