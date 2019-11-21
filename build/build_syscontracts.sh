#!/usr/bin/env bash

ABI=.cpp.abi.json
WASM=.wasm

syscontracts=(cnsManager cnsProxy nodeManager nodeRegister paramManager roleManager roleRegister userManager userRegister)

root=`pwd`
SYSCONTRACTS=$root/cmd/SysContracts
WORKSPACE_PATH=$root/release/linux/


cd $SYSCONTRACTS
rm -rf ${SYSCONTRACTS}/build/systemContract/*/*json ${SYSCONTRACTS}/build/systemContract/*/*wasm ${WORKSPACE_PATH}/conf/contracts
echo "remove sys abi and bytescode before rebuild"

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

cp ${SYSCONTRACTS}/build/systemContract/*/*json ${SYSCONTRACTS}/build/systemContract/*/*wasm  ${WORKSPACE_PATH}/conf/contracts/
echo "cp sys abi and bytescode files to $root/release/linux/conf/contracts"

echo "build system contracts successful"