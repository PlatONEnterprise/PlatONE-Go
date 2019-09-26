userAddr_string=$1

config=../../conf/ctool.json
addr=0x6d5ca050397a41eb1157b9ae61911cb4121e05b5
abi=../../conf/contracts/userManager.cpp.abi.json

../ctool invoke --config $config --addr $addr --abi $abi --func delUser --param $userAddr_string 
echo "delUser"
echo userAddr_string = $userAddr_string
