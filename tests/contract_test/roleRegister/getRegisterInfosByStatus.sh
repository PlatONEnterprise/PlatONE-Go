#!/bin/bash
cfg='cfg'
config=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
abi=`cat $cfg | grep ABI_JSON | awk -F'=' '{print $2}'`
addr=`cat $cfg | grep ADDR | awk -F'=' '{print $2}'`
ctool=`cat $cfg | grep CTOOL_BIN | awk -F'=' '{print $2}'`

status_int32=$1
pageNum_int32=$2
pageSize_int32=$3

$ctool invoke --config $config --addr $addr --abi $abi --func getRegisterInfosByStatus --param $status_int32 --param $pageNum_int32 --param $pageSize_int32 
echo "getRegisterInfosByStatus"
echo status_int32 = $status_int32
echo pageNum_int32 = $pageNum_int32
echo pageSize_int32 = $pageSize_int32
