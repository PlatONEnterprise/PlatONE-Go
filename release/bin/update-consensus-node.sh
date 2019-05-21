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

name_string="nodeA"
nodeJsonStr_string='{"type":1}'

./ctool invoke --config $config --addr $addr --abi $abi --func update --param $name_string --param $nodeJsonStr_string 
echo "update"
echo name_string = $name_string
echo nodeJsonStr_string = $nodeJsonStr_string

name_string="nodeB"
nodeJsonStr_string='{"type":1}'

./ctool invoke --config $config --addr $addr --abi $abi --func update --param $name_string --param $nodeJsonStr_string 
echo "update"
echo name_string = $name_string
echo nodeJsonStr_string = $nodeJsonStr_string

name_string="nodeC"
nodeJsonStr_string='{"type":1}'

./ctool invoke --config $config --addr $addr --abi $abi --func update --param $name_string --param $nodeJsonStr_string 
echo "update"
echo name_string = $name_string
echo nodeJsonStr_string = $nodeJsonStr_string
