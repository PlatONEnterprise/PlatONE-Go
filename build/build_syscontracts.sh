#!/usr/bin/env bash

ABI=.cpp.abi.json
WASM=.wasm

syscontracts=(cnsManager cnsProxy nodeManager nodeRegister paramManager roleManager roleRegister userManager userRegister)

root=`pwd`
SYSCONTRACTS=$root/cmd/SysContracts
WORKSPACE_PATH=$root/release/linux/


cd $SYSCONTRACTS
#for str in ${syscontracts[@]};do
#    find . -name $str$WASM -exec rm -rf {} \;
rm -rf ${SYSCONTRACTS}/build/systemContract/*/*json ${SYSCONTRACTS}/build/systemContract/*/*wasm ${WORKSPACE_PATH}/conf/contracts
echo "remove sys abi and bytescode before rebuild"
#    echo "remove $str$WASM before rebuild"

#    find . -name $str$ABI -exec rm -rf {} \;
#    echo "remove $str$ABI before rebuild"
#done

./script/build_system_contracts.sh

for str in ${syscontracts[@]};do
    if [ ! -f $SYSCONTRACTS/build/systemContract/$str/$str$WASM ]; then
        echo "not found $SYSCONTRACTS/build/systemContract/$str/$str$WASM"
        exit 1
    fi

    if [ ! -f $SYSCONTRACTS/build/systemContract/$str/$str$ABI ]; then
            echo "not found $SYSCONTRACTS/build/systemContract/$str/$str$ABI"
            exit 1
        fi
done

mkdir $root/release/linux/conf/contracts

#for str in ${syscontracts[@]};do
#    find . -name $str$WASM -exec cp {} $root/release/linux/conf/contracts \;
#    echo "cp $str$WASM file to $root/release/linux/conf/contracts"
#
#    find . -name $str$ABI -exec cp {} $root/release/linux/conf/contracts \;
#    echo "cp $str$ABI file to $root/release/linux/conf/contracts"
#done
cp ${SYSCONTRACTS}/build/systemContract/*/*json ${SYSCONTRACTS}/build/systemContract/*/*wasm  ${WORKSPACE_PATH}/conf/contracts/
echo "cp sys abi and bytescode files to $root/release/linux/conf/contracts"

echo "build system contracts successful"