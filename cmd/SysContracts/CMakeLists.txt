cmake_minimum_required( VERSION 2.8 )

project( BCWasm )
set(DEFAULT_SYSTEM_INCLUDE_FOLDERS ${CMAKE_SOURCE_DIR}/bcwasmlib/include ${CMAKE_SOURCE_DIR}/libc++/libcxx/include ${CMAKE_SOURCE_DIR}/libc/musl/include ${CMAKE_SOURCE_DIR}/boost/include ${CMAKE_SOURCE_DIR}/rapidjson/include)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
list(APPEND CMAKE_MODULE_PATH ${PROJECT_SOURCE_DIR}/CMakeModules)
set(libc ${CMAKE_SOURCE_DIR}/lib/libc.bc)
set(libc++ ${CMAKE_SOURCE_DIR}/lib/libc++.bc)
set(libbcwasm ${CMAKE_SOURCE_DIR}/lib/libbcwasm.bc)
list(APPEND DEFAULT_SYSTEM_INCLUDE_FOLDERS ${CMAKE_SOURCE_DIR}/external/lib/clang/4.0.1/include)

include(FindWasm)
include(wasm)


#add_subdirectory(libc++)
#add_subdirectory(libc)
#add_subdirectory(bcwasmlib)

if ( $ENV{TESTCONTRACT} )
    add_subdirectory(testContract)
elseif ( $ENV{TESTS} )
    add_subdirectory(unitTest)
elseif ( $ENV{APP} )
    add_subdirectory(appContract)  
else()
    add_subdirectory(systemContract)
endif()

