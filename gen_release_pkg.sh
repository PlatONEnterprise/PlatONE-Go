#!/bin/bash


Version=${1}
[[ ${Version} == "" ]] && echo "ERROR: Please set release version; example command:
     \"${0} v0.0.0.0\"" && exit

cd ..
PlatONE_Project_name="PlatONE-Go"
Java_Project_name="java-sdk"
BCWasm_project_name="BCWasm"

PlatONE_Linux_Dir="${PlatONE_Project_name}/release/linux"
PlatONE_CMD_SystemContract="${PlatONE_Project_name}/cmd/SysContracts"

PlatONE_linux_name="PlatONE_linux_${Version}"
BCWasm_linux_name="BCWasm_linux_release.${Version}"
Java_sdk_linux_name="java_sdk_linux_${Version}"
End_with=".tar.gz"

release_md_name="release.md"

function create_release_note() {
cat <<EOF
链部署指南
[PlatONE快速搭链教程](https://180.167.100.189:20443/PlatONE/doc/PlatONE_WIKI/blob/v0.9.0/zh-cn/basics/Installation/%5BChinese-Simplified%5D-%E5%BF%AB%E9%80%9F%E9%83%A8%E7%BD%B2.md)
->上传${PlatONE_linux_name}

WASM合约开发库
[PlatONE合约指导文档](https://180.167.100.189:20443/PlatONE/doc/PlatONE_WIKI/blob/v0.9.0/zh-cn/WASMContract/%5BChinese-Simplified%5D-%E5%90%88%E7%BA%A6%E6%95%99%E7%A8%8B.md)
->上传${BCWasm_linux_name}

SDK工具
[SDK使用说明](https://180.167.100.189:20443/PlatONE/doc/PlatONE_WIKI/blob/v0.9.0/zh-cn/SDK/%5BChinese-Simplified%5D-SDK%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E.md)
->上传${Java_sdk_linux_name}

Release Change Log
[change_log](https://180.167.100.189:20443/PlatONE/src/node/PlatONE-Go/blob/develop/CHANGELOG.md)
EOF
}

function env() {
    if [[ -d ${PlatONE_Project_name} ]]; then
        echo "${PlatONE_Project_name} already exists."
    else
        git clone --recursive https://172.16.211.192/PlatONE/src/node/PlatONE-Go.git
    fi

    if [[ -d ${Java_Project_name} ]]; then
        echo "${Java_Project_name} already exists"
    else
        git clone --recursive https://172.16.211.192/PlatONE/src/node/java-sdk.git
    fi
    rm -rf ${Java_Project_name}/.git
}

function compile() {
    cd ${PlatONE_Project_name} && make all && cd ..
}

function create_platone_linux() {
    [[ -d ${PlatONE_linux_name} ]] && rm -rf ${PlatONE_linux_name}
    mkdir ${PlatONE_linux_name}
    cp -rf ${PlatONE_Linux_Dir}/* ${PlatONE_linux_name}/
    tar -zcvf ${PlatONE_linux_name}${End_with} ${PlatONE_linux_name}
}

function create_bcwasm_linux() {
    [[ -d ${BCWasm_project_name} ]] && rm -rf ${BCWasm_project_name}
    mkdir ${BCWasm_project_name}
    cp -rf ${PlatONE_CMD_SystemContract}/* ${BCWasm_project_name}/
    cp ${PlatONE_Project_name}/release/linux/bin/ctool ${BCWasm_project_name}/external/bin/
    rm -rf ${BCWasm_project_name}/systemContract
    rm -rf ${BCWasm_project_name}/build
    tar -zcvf ${BCWasm_linux_name}${End_with} ${BCWasm_project_name}
}

function create_sdk_linux() {
    tar -zcvf ${Java_sdk_linux_name}${End_with} ${Java_Project_name}
}

function tag() {
    cd ${PlatONE_Project_name}
    git tag -a ${Version} -m "Release"
    git push
    cd ..
}

function clean() {
    rm -rf ${PlatONE_linux_name}
    rm -rf ${Java_Project_name}
    rm -rf ${BCWasm_project_name}
}

function main() {
    echo "#################################################################################"
    echo "note: Please change the version number in PlatONE-Go before executing this script"
    echo "#################################################################################"
    sleep 3
    echo "#################################################################################"
    echo "note: If it is github, please set the change log differently"
    echo "#################################################################################"
    sleep 3

    env
    compile

    create_platone_linux
    create_bcwasm_linux
    create_sdk_linux

    tag

    clean
    echo "#################################################################################"
    echo "note: The release pkg massage format:"
    echo "#################################################################################"
    create_release_note
}

main