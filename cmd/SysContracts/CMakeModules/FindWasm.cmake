# - Try to find WASM

# TODO: Check if compiler is able to generate wasm32
if ("${CLANG_ROOT}" STREQUAL "")
    if (APPLE)
        set(CLANG_ROOT "/usr" )
    elseif (UNIX AND NOT APPLE)
        set(CLANG_ROOT "/usr" )
    elseif(WIN32)
        message(FATAL_ERROR "CLANG don't know where to look please specify CLANG_ROOT")
    else()
       message(FATAL_ERROR "CLANG not found and don't know where to look, please specify CLANG_ROOT")
   endif()
endif()

if ("${WASM_ROOT}" STREQUAL "")
    if (APPLE)
        set( WASM_ROOT "/usr" )
    elseif (UNIX AND NOT APPLE)
        set( WASM_ROOT "/usr" )
    elseif(WIN32)
        message(FATAL_ERROR "WASM_ROOT don't know where to look please specify CLANG_ROOT")
    else()
      message(FATAL_ERROR "WASM not found and don't know where to look, please specify WASM_ROOT")
   endif()
endif()

if ("${BCWASM_TOOL_ROOT}" STREQUAL "")
    if (APPLE)
        set(BCWASM_TOOL_ROOT "/usr" )
    elseif (UNIX AND NOT APPLE)
        set(BCWASM_TOOL_ROOT "/usr" )
    elseif(WIN32)
        message(FATAL_ERROR "BCWASM_TOOL_ROOT don't know where to look please specify CLANG_ROOT")
    else()
       message(FATAL_ERROR "BCWASM not found and don't know where to look, please specify BCWASM_TOOL_ROOT")
   endif()
endif()

if (WIN32)
    find_program(WASM_CLANG clang.exe PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(WASM_LLC llc.exe PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(WASM_LLVM_LINK llvm-link.exe PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-S2WASM bcwasm-s2wasm.exe PATHS ${WASM_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-WAST2WASM bcwasm-wast2wasm.exe PATHS ${WASM_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-ABIGEN bcwasm-abigen.exe PATHS ${BCWASM_TOOL_ROOT}/bin NO_DEFAULT_PATH)
else()
    find_program(WASM_CLANG clang PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(WASM_LLC llc PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(WASM_LLVM_LINK llvm-link PATHS ${CLANG_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-S2WASM bcwasm-s2wasm PATHS ${WASM_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-WAST2WASM bcwasm-wast2wasm PATHS ${WASM_ROOT}/bin NO_DEFAULT_PATH)
    find_program(BCWASM-ABIGEN bcwasm-abigen PATHS ${BCWASM_TOOL_ROOT}/bin NO_DEFAULT_PATH)
endif()


include(FindPackageHandleStandardArgs)
# handle the QUIETLY and REQUIRED arguments and set EOS_FOUND to TRUE
# if all listed variables are TRUE

find_package_handle_standard_args(WASM REQUIRED_VARS WASM_CLANG WASM_LLC WASM_LLVM_LINK)

