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
#ccount="dc22bea8701610366d4b62a40a9130731372f56d"

# get node pubkey


publicKey=`cat ../../data/node.pubkey`
echo "plase input root node ip"
read ip
nodeJsonStr_string="{\"name\":\"${nodeName}\",\"type\":1,\"publicKey\":\"${publicKey}\",\"desc\":\"desc\",\"externalIP\":\"${ip}\",\"internalIP\":\"${ip}\",\"rpcPort\":6791,\"p2pPort\":16791,\"root\":true,\"owner\":\"0x${account}\",\"status\":1}"

./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
echo "add"
echo nodeJsonStr_string = $nodeJsonStr_string


list=`cat ../conf/node_list`

count=1
for i in ${list[*]}
do

  for j in {1..1}
  do
    [ ${count} -eq 1 ] && ((count=count+1)) &&  continue
    nodeName=node${count}
    publicKey=`cat ../../data${count}/node.pubkey`
    externalIP=${i}
    nodeJsonStr_string="{\"name\":\"${nodeName}\",\"type\":0,\"publicKey\":\"${publicKey}\",\"desc\":\"desc\",\"externalIP\":\"${externalIP}\",\"internalIP\":\"${externalIP}\",\"rpcPort\":679${j},\"p2pPort\":1679${j},\"root\":false,\"owner\":\"0x${account}\",\"status\":1}"

    ./ctool invoke --config $config --addr $addr --abi $abi --func add --param $nodeJsonStr_string 
    echo "add"
    echo nodeJsonStr_string = $nodeJsonStr_string
    ((count=count+1))
    echo ${count}
  done
done
