#!/bin/sh

pids=`ps -ef | grep bcos | grep -v grep | awk '{print $2}'`

if [ $pids"x" != "x" ]; then
    echo "An old bcos process[$pids] is runing, please stop it first."
    exit
fi

#mkdir -p ../logs

datadir="--datadir ../data"
nodekey="--nodekey ../data/node.prikey"
rpc="--rpcaddr 0.0.0.0 --rpcport 6789 --rpcapi db,eth,net,web3,admin,personal --rpc"
#logs="--verbosity 4 --wasmlog ../logs/wasm.log >>../logs/bcos.log" #redirection not work in scrypt, why?
logs="--verbosity 4 --wasmlog ./wasm.log"

nohup ./bcos --identity bcos --nodiscover --rpccorsdomain "*" $datadir $nodekey $rpc $logs 2>&1 &

sleep 1

pids=`ps -ef | grep bcos | grep -v grep | awk '{print $2}'`

if [ $pids"x" != "x" ]; then
    echo "Start bcos succ. pid[$pids]"
fi
