#!/bin/bash

createDirectory(){
    if [ -d ${TARGET_PATH} ]
    then
        echo "had create dir ${TARGET_PATH}"
        exit 1
    fi

    mkdir -p ${TARGET_PATH}
    echo -e "add_wast_executable(TARGET ${TARGET} \n \
    #Multiple cpp files use SOURCE_FILES, and the name of TARGET is the same as the cpp file name with BCWASM_ABI \n \
    #SOURCE_FILES multi filenames \n \
    INCLUDE_FOLDERS "\${STANDARD_INCLUDE_FOLDERS}" \n \
    LIBRARIES  \${libbcwasm} \${libc++} \${libc} \n \
    DESTINATION_FOLDER \${CMAKE_CURRENT_BINARY_DIR})" >> ${TARGET_PATH}/CMakeLists.txt

    echo "//auto create contract
#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

namespace demo {
class my_contract : public bcwasm::Contract
{
    public:
    my_contract(){}

    /// 实现父类: bcwasm::Contract 的虚函数
    /// 该函数在合约首次发布时执行，仅调用一次
    void init() 
    {
        bcwasm::println(\"init success...\");
    }
    /// 定义Event.
    /// BCWASM_EVENT(eventName, arguments...)
    BCWASM_EVENT(setName, const char *)
    
    public:
    void setName(const char *msg)
    {
        // 定义状态变量
        bcwasm::setState(\"NAME_KEY\", std::string(msg));
        // 日志输出
        // 事件返回
        BCWASM_EMIT_EVENT(setName, \"std::string(msg)\");
    }
    const char* getName() const 
    {
        std::string value;
        bcwasm::getState(\"NAME_KEY\", value);
        // 读取合约数据并返回
        return value.c_str();
    }
};
}
// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::my_contract, setName)
BCWASM_ABI(demo::my_contract, getName)

//bcwasm autogen end    
        
    " > "${TARGET_PATH}/${TARGET}.cpp"
}

appendCMakeDir(){
    cmd="add_subdirectory(${TARGET})"
    add=`cat ${USER_PATH}/CMakeLists.txt | grep "${cmd}" | wc -l`

    if [ "-$add" -eq "-0" ]
    then
        echo -e "\n${cmd}" >> "${USER_PATH}/CMakeLists.txt"
    else
        echo "find it !!!"
    fi
}

build(){
    builddir=${HOME}/build
    externaldir=${HOME}/external
    if [ ! -d ${builddir} ]
    then
        mkdir ${builddir}
    fi
    cd $builddir && cmake .. -DWASM_ROOT=$externaldir -DBCWASM_TOOL_ROOT=$externaldir -DCLANG_ROOT=$externaldir && make
}

userBuild(){
    builddir=${HOME}/build
    externaldir=${HOME}/external
    if [ ! -d ${builddir} ]
    then
        mkdir ${builddir}
    fi
    cd $builddir && cmake .. -DWASM_ROOT=$externaldir -DBCWASM_TOOL_ROOT=$externaldir -DCLANG_ROOT=$externaldir && make
}

main(){

    if [ "${HOME}" == "help" ]
    then
        echo "autoproject [home] [target]"
        exit 1
    fi
   
    if [[ "-" == "-${HOME}" ]]  || [[ "-." == "-${HOME}" ]]
    then
        HOME=`pwd`
    fi

    if [ ! -d "${HOME}" ]
    then
        echo "doesn't have ${HOME}"
        exit 1
    fi
    
  

    if [ "-${TARGET}" == "-" ]
    then
        build
        #cp ${BIN_PATH}/ctool ${SCRIPT_DIR}/../build
        cp ../external/bin/ctool ${SCRIPT_DIR}/../build
    else 
        createDirectory   

        if [ "${SIGN}" == "vc" ]
        then
            cp ${USER_PATH}/vccMain.cpp ${TARGET_PATH}/${TARGET}.cpp
        fi

        num=$?
        if [ -n $num ]
        then
            appendCMakeDir
        fi
        userBuild
    fi
}

# variables
SCRIPT_NAME=$0
SCRIPT_DIR_R=`dirname "${SCRIPT_NAME}"`
CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR_R}
SCRIPT_DIR=`pwd`
cd ${CURRENT_PATH}


CURRENT_PATH=`pwd`
cd ${SCRIPT_DIR}/../../..
WORKSPACE_PATH=`pwd`  # PlatONE-Go/
cd ${CURRENT_PATH}

BIN_PATH=${WORKSPACE_PATH}/release/linux/bin


# HOME=${WORKSPACE_PATH}/cmd/SysContracts # PlatONE-Go/cmd/SystemContracts
HOME=.
TARGET=$2
SIGN=$3
USER_PATH="${HOME}/appContract"
TARGET_PATH="${USER_PATH}/$2"
export APP=1

main

