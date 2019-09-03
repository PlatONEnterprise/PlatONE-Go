#!/bin/bash

HOME=$1
TARGET=$2
SIGN=$3
USER_PATH="${HOME}/test"
TARGET_PATH="${USER_PATH}/$2"
export TESTS=2

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

    echo "//auto create contract" > "${TARGET_PATH}/${TARGET}.cpp"
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
    builddir=${HOME}/buildUnitTest
    externaldir=${HOME}/external
    if [ ! -d ${builddir} ]
    then
        mkdir ${builddir}
    fi
    cd $builddir && cmake .. -DWASM_ROOT=$externaldir -DBCWASM_TOOL_ROOT=$externaldir -DCLANG_ROOT=$externaldir && make
}

userBuild(){
    builddir=${HOME}/buildUnitTest
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

main

