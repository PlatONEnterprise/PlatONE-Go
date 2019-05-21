#!/bin/bash


./stop.sh

./init_four.sh

#cd /home/wxuser/platone_cluster_line2/ &&./reset.sh
#cd -

./one_click_four.sh

echo "input nodeManagerAdd:"
read nodeManagerAdd
echo "input owner:"
read owner
sleep 60

key=`cat ../../data5/node.pubkey`

./ctool invoke --config ../conf/ctool.json --addr ${nodeManagerAdd} --abi ../conf/contracts/nodeManager.cpp.abi.json --func add --param '{"name":"node5","type":0,"publicKey":"'${key}'","desc":"desc","externalIP":"127.0.0.1","internalIP":"127.0.0.1","rpcPort":6795,"p2pPort":16795,"root":false,"owner":"'${owner}'","status":1}'

/home/wxuser/platone_cluster_line2/platone --identity node5 --nodiscover --datadir /home/wxuser/platone_cluster_line2/data5 --nodekey /home/wxuser/platone_cluster_line2/data5/node.prikey --rpcaddr 0.0.0.0 --rpcport 6795 --rpcapi db,eth,net,web3,admin,personal --rpc --verbosity 4 --wasmlog /home/wxuser/platone_cluster_line2/data5/wasm.log --port 16795 --ws --wsport 26795 > /home/wxuser/platone_cluster_line2/data5/platone.log 2>&1 &
