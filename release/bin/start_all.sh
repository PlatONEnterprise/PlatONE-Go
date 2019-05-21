#!/bin/sh

pids=`ps -ef | grep platone | grep -v grep | awk '{print $2}'`

#if [ $pids"x" != "x" ]; then
#    echo "old platone process[$pids] is runing, please stop it first."
#    exit
#fi

cd ../..
workdir=`pwd`
cd -
bindir=`pwd`

#mkdir -p ../logs

datadir="--datadir ../../data"
datadirA="--datadir ../../dataA"
datadirB="--datadir ../../dataB"
datadirC="--datadir ../../dataC"

nodekey="--nodekey ../../data/node.prikey"
nodekeyA="--nodekey ../../dataA/node.prikey"
nodekeyB="--nodekey ../../dataB/node.prikey"
nodekeyC="--nodekey ../../dataC/node.prikey"

rpc="--rpcaddr 0.0.0.0 --rpcport 6791 --rpcapi db,eth,net,web3,admin,personal --rpc"
rpcA="--rpcaddr 0.0.0.0 --rpcport 6792 --rpcapi db,eth,net,web3,admin,personal --rpc"
rpcB="--rpcaddr 0.0.0.0 --rpcport 6793 --rpcapi db,eth,net,web3,admin,personal --rpc"
rpcC="--rpcaddr 0.0.0.0 --rpcport 6794 --rpcapi db,eth,net,web3,admin,personal --rpc"

#logs="--verbosity 4 --wasmlog ../logs/wasm.log >>../logs/platon.log" #redirection not work in scrypt, why?
logs="--verbosity 3 --wasmlog ../../data/wasm.log"
logsA="--verbosity 4 --wasmlog ../../dataA/wasm.log"
logsB="--verbosity 4 --wasmlog ../../dataB/wasm.log"
logsC="--verbosity 4 --wasmlog ../../dataC/wasm.log"

#nohup ./platone --identity platone --nodiscover $datadir $nodekey $rpc $logs >platone.log 2>&1 &
#nohup ./platone --identity node --nodiscover $datadir $nodekey $rpc $logs --port 16791 --ws --wsport 26791 > ../../data/platone.log 2>&1 & 
#echo $!

#nohup ./platone --identity platone --nodiscover $datadirA $nodekeyA $rpcA $logsA --port 16792 > ../../dataA/platone.log 2>&1 & 
#echo $!

#nohup ./platone --identity platone --nodiscover $datadirB $nodekeyB $rpcB $logsB --port 16793 > ../../dataB/platone.log 2>&1 & 
#echo $!

#nohup ./platone --identity platone --nodiscover $datadirC $nodekeyC $rpcC $logsC --port 16794 > ../../dataC/platone.log 2>&1 & 
#echo $!

# sleep 1


list=`cat ../conf/node_list`

count=1
for i in ${list[*]}
do
  for j in {1..1};do
    [ ${count} -eq 1 ] && ((count=count+1)) &&  continue
    datadir="--datadir ${workdir}/data"${count}
    nodekey="--nodekey ${workdir}/data${count}/node.prikey"
    rpc="--rpcaddr 0.0.0.0 --rpcport 679${j} --rpcapi db,eth,net,web3,admin,personal --rpc"
    logs="--verbosity 4 --wasmlog ${workdir}/data${count}/wasm.log"
    if [ $count -lt 11 ];then
      ssh $i "${bindir}/platone --identity node${count} --nodiscover $datadir $nodekey $rpc $logs --port 1679${j} --ws --wsport 2679${j} --gcmode archive > ${workdir}/data${count}/platone.log 2>&1 &"
    else
      ssh $i "${bindir}/platone --identity node${count} --nodiscover $datadir $nodekey $rpc $logs --port 1679${j} --ws --wsport 2679${j} --rpccorsdomain \"*\"> ${workdir}/data${count}/platone.log 2>&1 &"
    fi
    
    echo $!
    ((count=count+1))
    echo ${count}
  done
done

# pids=`ps -ef | grep platone | grep -v grep | awk '{print $2}'`

# if [ $pids"x" != "x" ]; then
#     echo "Start platone succ. pid[$pids]"
# fi
