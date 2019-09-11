#!/bin/bash
log=./tmp
cfg='cfg'
ctool_json=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
wasm_log=`cat $cfg | grep WASM_LOG | awk -F'=' '{print $2}'`

root_addr=`cat $ctool_json | grep from |awk -F':' '{print $2}' | sed 's/\"//g'`


runTest() {
  echo "START RUNNING CASE FOR $cmd"
  echo
  count=0
  for i in ${params[*]}; do
    param=$i
    if [ "$i" != "" ]; then
       param=`echo $i|sed 's/|/ /g'`
    fi
    echo "Case ${count}:"
    echo "input param is: " $param
    old_log=`tail -n3 ${wasm_log}`
    $cmd $param > $log 
    expect=${expect_code[$count]}
    cat $log | grep -a code
    runcode=`cat $log | grep -a code |awk -F':' '{print $3}'|awk -F',' '{print $1}'`
    if [ "$runcode" == "" ];then 
      runcode=`cat $log |grep -a ^result| awk -F':' '{print $2}' | sed 's/^[ \t]*//g'` 
      [ "$runcode" != "" ] && ((runcode=runcode-1)) && runcode=${runcode#-}
    fi
    if [ "$runcode" != "" ]&&[ "$expect" != "" ]; then 
      [ "$expect"  == "$runcode" ] && echo -e "\033[32m [ OK ] \033[0m" || echo -e "\033[31m [ FAILED ] \033[0m"
    fi
    sleep 2
    new_log=`tail -n3 $wasm_log`
    [ "$old_log" != "$new_log" ] && echo "wasm.log: " &&  tail -n3 $wasm_log && echo

    ((count=count+1))
    
  done
  echo "---------------------------------"
  echo && echo
}

#############
# START TO MODIFY HERE

# getAccountByName
params=('root' 'aaa' \'\')
expect_code=(0 1 1)
cmd=./getAccountByName.sh
runTest

# getAccountByAddress
params=('0x5cb3c740bf48512ae2cd67fb766bf0883ad1fd3b' '0x11222' \'\' $root_addr)
expect_code=(1 1 1 0)
cmd=./getAccountByAddress.sh
runTest

#############

# addUser
#user1='{"address":"0x11","name":"xiaoluo","mobile":"131111","email":"11@qq.com","status":0}'
#params=($user1)
#expect_code=(0)
#cmd=./addUser.sh
#runTest

#############
#isValidUser
params=('0x5cb3c740bf48512ae2cd67fb766bf0883ad1fd3b' $root_addr \'\')
expect_code=(1 0 1)
cmd=./isValidUser.sh
runTest

#############
#enable
params=($root_addr '0x11' \'\')
expect_code=(0 1 1)
cmd=./enable.sh
runTest

#############
#disable
params=('\"b3c740bf48512ae2cd67fb766bf0883ad1fd3b\"' '0x11' \'\')
expect_code=(1 1 1)
cmd=./disable.sh
runTest

#############
#update
params=("${root_addr}|{"mobile":"132","email":"126qq.com","status":0}")
expect_code=(0)
cmd=./update.sh
runTest 2


rm $log
