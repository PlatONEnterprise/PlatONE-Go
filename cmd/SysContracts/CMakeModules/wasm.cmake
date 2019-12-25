find_package(Wasm)

if(WASM_FOUND)
  message(STATUS "Using WASM clang => " ${WASM_CLANG})
  message(STATUS "Using WASM llc => " ${WASM_LLC})
  message(STATUS "Using WASM llvm-link => " ${WASM_LLVM_LINK})
else()
  message( FATAL_ERROR "No WASM compiler cound be found (make sure WASM_ROOT is set)" )
  return()
endif()
macro(compile_wast)
  #read arguments include ones that we don't since arguments get forwared "as is" and we don't want to threat unknown argument names as values
  cmake_parse_arguments(ARG "NOWARNINGS" "TARGET;DESTINATION_FOLDER" "SOURCE_FILES;INCLUDE_FOLDERS;SYSTEM_INCLUDE_FOLDERS;LIBRARIES" ${ARGN})
  set(target ${ARG_TARGET})

  # NOTE: Setting SOURCE_FILE and looping over it to avoid cmake issue with compilation ${target}.bc's rule colliding with
  # linking ${target}.bc's rule
  if ("${ARG_SOURCE_FILES}" STREQUAL "")
    set(SOURCE_FILES ${target}.cpp)
  else()
    set(SOURCE_FILES ${ARG_SOURCE_FILES})
  endif()
  set(outfiles "")
  foreach(srcfile ${SOURCE_FILES})

    get_filename_component(outfile ${srcfile} NAME)
    get_filename_component(extension ${srcfile} EXT)
    get_filename_component(infile ${srcfile} ABSOLUTE)

    # -ffreestanding
    #   Assert that compilation targets a freestanding environment.
    #   This implies -fno-builtin. A freestanding environment is one in which the standard library may not exist, and program startup may not necessarily be at main.
    #   The most obvious example is an OS kernel.

    # -nostdlib
    #   Do not use the standard system startup files or libraries when linking.
    #   No startup files and only the libraries you specify are passed to the linker, and options specifying linkage of the system libraries, such as -static-libgcc or -shared-libgcc, are ignored.
    #   The compiler may generate calls to memcmp, memset, memcpy and memmove.
    #   These entries are usually resolved by entries in libc. These entry points should be supplied through some other mechanism when this option is specified.

    # -fno-threadsafe-statics
    #   Do not emit the extra code to use the routines specified in the C++ ABI for thread-safe initialization of local statics.
    #   You can use this option to reduce code size slightly in code that doesn’t need to be thread-safe.

    # -fno-rtti
    #   Disable generation of information about every class with virtual functions for use by the C++ run-time type identification features (dynamic_cast and typeid).

    # -fno-exceptions
    #   Disable the generation of extra code needed to propagate exceptions
    if ("${extension}" STREQUAL ".c")
      set(STDFLAG -D_XOPEN_SOURCE=700)
    else()
      set(STDFLAG "--std=c++14")
    endif()

    set(WASM_COMMAND ${WASM_CLANG} -emit-llvm -O3 ${STDFLAG} --target=wasm32
      -nostdlib -nostdlibinc -DBOOST_DISABLE_ASSERTS -DBOOST_EXCEPTION_DISABLE -DWASM_COMPILE -fno-threadsafe-statics -fno-rtti -fno-exceptions -w
      -c ${infile} -o ${outfile}.bc
      )

    # rapidjson使用64位结构,解决double精度丢失问题
    list(APPEND WASM_COMMAND -DRAPIDJSON_64BIT=1)

    if (${ARG_NOWARNINGS})
      list(APPEND WASM_COMMAND -Wno-everything)
    else()
      list(APPEND WASM_COMMAND -Weverything -Wno-c++98-compat -Wno-old-style-cast -Wno-vla -Wno-vla-extension -Wno-c++98-compat-pedantic
        -Wno-missing-prototypes -Wno-missing-variable-declarations -Wno-packed -Wno-padded -Wno-c99-extensions  -Wno-documentation-unknown-command)
    endif()

    foreach(folder ${ARG_INCLUDE_FOLDERS})
      list(APPEND WASM_COMMAND -I ${folder})
    endforeach()

    if ("${ARG_SYSTEM_INCLUDE_FOLDERS}" STREQUAL "")
      set (ARG_SYSTEM_INCLUDE_FOLDERS ${DEFAULT_SYSTEM_INCLUDE_FOLDERS})
    endif()
    foreach(folder ${ARG_SYSTEM_INCLUDE_FOLDERS})
      list(APPEND WASM_COMMAND -isystem ${folder})
    endforeach()


    add_custom_command(OUTPUT ${outfile}.bc
      DEPENDS ${infile}
      COMMAND ${WASM_COMMAND}
      IMPLICIT_DEPENDS CXX ${infile}
      COMMENT "Building LLVM bitcode ${outfile}.bc"
      WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
      VERBATIM
      )
    set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${outfile}.bc)



    list(APPEND outfiles ${outfile}.bc)

  endforeach(srcfile)

  set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${target}.bc)

endmacro(compile_wast)

macro(add_wast_library)
  cmake_parse_arguments(ARG "NOWARNINGS" "TARGET;DESTINATION_FOLDER" "SOURCE_FILES;INCLUDE_FOLDERS;SYSTEM_INCLUDE_FOLDERS" ${ARGN})
  set(target ${ARG_TARGET})
  compile_wast(${ARGV})

  get_filename_component("${ARG_TARGET}_BC_FILENAME" "${ARG_DESTINATION_FOLDER}/${ARG_TARGET}.bc" ABSOLUTE CACHE)
  add_custom_target(${target} ALL DEPENDS ${${ARG_TARGET}_BC_FILENAME})

  add_custom_command(OUTPUT ${${ARG_TARGET}_BC_FILENAME}
    DEPENDS ${outfiles}
    COMMAND ${WASM_LLVM_LINK} -o ${${ARG_TARGET}_BC_FILENAME} ${outfiles}
    COMMENT "Linking LLVM bitcode library ${target}.bc"
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    VERBATIM
    )

endmacro(add_wast_library)





macro(compile_executable)
  cmake_parse_arguments(ARG "NOWARNINGS" "TARGET;DESTINATION_FOLDER;MAX_MEMORY;ABI" "SOURCE_FILES;INCLUDE_FOLDERS;SYSTEM_INCLUDE_FOLDERS;LIBRARIES" ${ARGN})
    set(target ${ARG_TARGET})
    set(DESTINATION_FOLDER ${ARG_DESTINATION_FOLDER})
    #message("compile_wast")

    if ("${ARG_SYSTEM_INCLUDE_FOLDERS}" STREQUAL "")
        set (ARG_SYSTEM_INCLUDE_FOLDERS ${DEFAULT_SYSTEM_INCLUDE_FOLDERS})
    endif()

    set(ABI_INCLUDE "")
    #message("INCLUDE_FOLDERS:${ARG_INCLUDE_FOLDERS}")
    foreach(folder ${ARG_INCLUDE_FOLDERS})
        list(APPEND ABI_INCLUDE "-extra-arg=-I${folder}")
    endforeach()

    foreach(folder ${ARG_SYSTEM_INCLUDE_FOLDERS})
        list(APPEND ABI_INCLUDE "-extra-arg=-I${folder}")
    endforeach()

    #message("ABI_INCLUDE ${ABI_INCLUDE}")
    if ("${ARG_SOURCE_FILES}" STREQUAL "")
        set(SOURCE_FILES ${target}.cpp)
    else()
        set(SOURCE_FILES ${ARG_SOURCE_FILES})
    endif()

    compile_wast(${ARGV})

    foreach(lib ${ARG_LIBRARIES})
        list(APPEND LIBRARIES ${lib})
    endforeach()
    #message("lib---${LIBRARIES} outfile :${outfile}")
    add_custom_command(OUTPUT ${target}.bc
      DEPENDS ${outfiles} ${ARG_LIBRARIES} ${LIBRARIES} ${ARG_ABI}
            COMMAND ${WASM_LLVM_LINK} -only-needed -o ${target}.bc ${outfiles} ${LIBRARIES}
            COMMENT "Linking LLVM bitcode executable ${target}.bc"
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            VERBATIM
            )

    set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${target}.bc)

    add_custom_command(OUTPUT ${target}.s
            DEPENDS ${target}.bc
            COMMAND ${WASM_LLC} -thread-model=single -asm-verbose=false -o ${target}.s ${target}.bc
            COMMENT "Generating textual assembly ${target}.s"
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            VERBATIM
            )
    set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${target}.s)

    if(ARG_MAX_MEMORY)
        set(MAX_MEMORY_PARAM "-m" ${ARG_MAX_MEMORY})
    endif()
    add_custom_target(${target} ALL DEPENDS   ${DESTINATION_FOLDER}/${target}.wasm)

    #由于binary 在windows执行成功后出现段错误，故采用执行脚本方式，强制返回成功
    if(WIN32)
        add_custom_command(OUTPUT ${DESTINATION_FOLDER}/${target}.wast
                DEPENDS ${target}.s
                COMMAND ${PROJECT_SOURCE_DIR}/script/runwast.bat ${BCWASM-S2WASM} ${DESTINATION_FOLDER}/${target}.wast  102400 ${DESTINATION_FOLDER}/${target}.s
                COMMENT "Generating WAST ${target}.wast"
                WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
                VERBATIM
                )
    else()
        add_custom_command(OUTPUT ${DESTINATION_FOLDER}/${target}.wast
                DEPENDS ${target}.s
                COMMAND ${BCWASM-S2WASM} -o ${DESTINATION_FOLDER}/${target}.wast -s 102400 ${MAX_MEMORY_PARAM} ${target}.s
                COMMENT "Generating WAST ${target}.wast"
                WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
                VERBATIM
                )
    endif()
    set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${target}.wast)

    if(WIN32)
        add_custom_command(OUTPUT ${DESTINATION_FOLDER}/${target}.wasm
                DEPENDS ${target}.wast
                COMMAND ${PROJECT_SOURCE_DIR}/script/runwasm.bat ${BCWASM-WAST2WASM} ${DESTINATION_FOLDER}/${target}.wast  ${DESTINATION_FOLDER}/${target}.wasm
                COMMENT "Generating WAST ${target}.wasm"
                WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
                VERBATIM
                )
    else()
        add_custom_command(OUTPUT ${DESTINATION_FOLDER}/${target}.wasm
                DEPENDS ${target}.wast
                COMMAND ${BCWASM-WAST2WASM} ${DESTINATION_FOLDER}/${target}.wast -o ${DESTINATION_FOLDER}/${target}.wasm
                COMMENT "Generating WASM ${target}.wasm"
                WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
                VERBATIM
                )
    endif()
    set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${target}.wasm)
endmacro(compile_executable)


macro(add_wast_executable)
    cmake_parse_arguments(ARG "NOWARNINGS" "TARGET;DESTINATION_FOLDER;MAX_MEMORY" "SOURCE_FILES;INCLUDE_FOLDERS;SYSTEM_INCLUDE_FOLDERS;LIBRARIES" ${ARGN})

    set(target ${ARG_TARGET})
    set(DESTINATION_FOLDER ${ARG_DESTINATION_FOLDER})
    #message("compile_wast")

    if ("${ARG_SYSTEM_INCLUDE_FOLDERS}" STREQUAL "")
        set (ARG_SYSTEM_INCLUDE_FOLDERS ${DEFAULT_SYSTEM_INCLUDE_FOLDERS})
    endif()

    set(ABI_INCLUDE "")
    #message("INCLUDE_FOLDERS:${ARG_INCLUDE_FOLDERS}")
    foreach(folder ${ARG_INCLUDE_FOLDERS})
        list(APPEND ABI_INCLUDE "-extra-arg=-I${folder}")
    endforeach()

    foreach(folder ${ARG_SYSTEM_INCLUDE_FOLDERS})
        list(APPEND ABI_INCLUDE "-extra-arg=-I${folder}")
    endforeach()

    #message("ABI_INCLUDE ${ABI_INCLUDE}")
    if ("${ARG_SOURCE_FILES}" STREQUAL "")
        set(SOURCE_FILES ${target}.cpp)
    else()
        set(SOURCE_FILES ${ARG_SOURCE_FILES})
    endif()

    set(abi "")

    foreach(srcfile ${SOURCE_FILES})
        get_filename_component(infile ${srcfile} ABSOLUTE)
        if (infile MATCHES ".*${target}.cpp.*")
            list(APPEND abi ${target}.cpp.abi.json)
            add_custom_target(${target}.cpp.abi.json ALL DEPENDS  ${DESTINATION_FOLDER}/${target}.cpp.abi.json)
            add_custom_command(OUTPUT ${target}.cpp.abi.json
                    DEPENDS ${infile}
                    PRE_BUILD
                    COMMAND ${BCWASM-ABIGEN} -extra-arg=-std=c++14  -extra-arg=--target=wasm32 -extra-arg=-w ${ABI_INCLUDE} ${infile} -- -outpath=${DESTINATION_FOLDER}
                    IMPLICIT_DEPENDS CXX ${infile}
                    COMMENT "Generating JSON ${infle}.abi.json"
                    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
                    VERBATIM
                    )

            set_property(DIRECTORY APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES ${infile}.abi.json)
        endif()
    endforeach()
    compile_executable(${ARGV} 
      ABI ${abi}
      )
endmacro(add_wast_executable)

macro(add_wast_test)
    cmake_parse_arguments(ARG "NOWARNINGS" "TARGET;DESTINATION_FOLDER;MAX_MEMORY" "SOURCE_FILES;INCLUDE_FOLDERS;SYSTEM_INCLUDE_FOLDERS;LIBRARIES" ${ARGN})
    compile_executable(${ARGV})
endmacro(add_wast_test)

