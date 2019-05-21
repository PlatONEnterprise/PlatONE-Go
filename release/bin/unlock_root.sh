read -p "Your account password?: " pw


# get node owner address
keyinfo=`cat ../../data/keystore/keyfile.json | sed s/[[:space:]]//g`
keyinfo=${keyinfo,,}sss
account=${keyinfo:12:40}

curl -X POST  -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"0x${account}\",\"${pw}\",0],\"id\":1}" http://localhost:6791