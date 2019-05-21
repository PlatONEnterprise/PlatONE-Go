config="../conf/ctool.json"
abi="../conf/contracts/nodeManager.cpp.abi.json"

# get nodeManager address
abiCNS="../conf/contracts/cnsManager.cpp.abi.json"
addrCNS="0x0000000000000000000000000000000000000011"
func="getContractAddress"
param1="__sys_NodeManager"
param2="latest"
ret=`./ctool invoke --config $config --abi $abiCNS --addr $addrCNS --func $func --param $param1 --param $param2 | sed s/[[:space:]]//g`

addr=${ret#*result:}

# get node owner address
keyinfo=`cat ../../data/keystore/keyfile.json | sed s/[[:space:]]//g`
keyinfo=${keyinfo,,}
account=${keyinfo:12:40}

# get node pubkey
publicKey=`cat ../../data/node.pubkey`
root_ip=`cat ../conf/node_list|awk '{print $1}'`

nodeJsonStr_string="{\"name\":\"node\",\"type\":1,\"publicKey\":\"${publicKey}\",\"desc\":\"desc\",\"externalIP\":\"${root_ip}\",\"internalIP\":\"${root_ip}\",\"rpcPort\":6791,\"p2pPort\":16791,\"root\":true,\"owner\":\"0x${account}\",\"status\":1}"

./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
echo "add"
echo nodeJsonStr_string = $nodeJsonStr_string

exit

nodeJsonStr_string="{\"name\":\"nodeA\",\"type\":0,\"publicKey\":\"${publicKeyA}\",\"desc\":\"desc\",\"externalIP\":\"127.0.0.1\",\"internalIP\":\"127.0.0.1\",\"rpcPort\":6792,\"p2pPort\":16792,\"root\":false,\"owner\":\"0x${account}\",\"status\":1}"

./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
echo "add"
echo nodeJsonStr_string = $nodeJsonStr_string

nodeJsonStr_string="{\"name\":\"nodeB\",\"type\":0,\"publicKey\":\"${publicKeyB}\",\"desc\":\"desc\",\"externalIP\":\"127.0.0.1\",\"internalIP\":\"127.0.0.1\",\"rpcPort\":6793,\"p2pPort\":16793,\"root\":false,\"owner\":\"0x${account}\",\"status\":1}"

./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
echo "add"
echo nodeJsonStr_string = $nodeJsonStr_string

nodeJsonStr_string="{\"name\":\"nodeC\",\"type\":0,\"publicKey\":\"${publicKeyC}\",\"desc\":\"desc\",\"externalIP\":\"127.0.0.1\",\"internalIP\":\"127.0.0.1\",\"rpcPort\":6794,\"p2pPort\":16794,\"root\":false,\"owner\":\"0x${account}\",\"status\":1}"

./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
echo "add"
echo nodeJsonStr_string = $nodeJsonStr_string


