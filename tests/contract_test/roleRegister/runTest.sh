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


params=([\"contractAdmin\"])
expect_code=(0)
cmd=./registerRole.sh 
runTest

params=($root_addr '0x111')
expect_code=(0 1)
cmd=./getRegisterInfoByAddress.sh
runTest

params=('root' '\"\"' 'somebody')
expect_code=(0 1 1)
cmd=./getRegisterInfoByName.sh
runTest

params=('1|0|10' '2|0|10' '0|0|10' '1|0|0')
expect_code=(0 1 1 0)
cmd=./getRegisterInfosByStatus.sh
runTest

params=("${root_addr}|3")
expect_code=(1)
cmd=./approveRole.sh
runTest






rm $log
