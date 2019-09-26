file(GLOB TESTCASE_SOURCE  *.cpp)

foreach(srcfile ${TESTCASE_SOURCE})
    get_filename_component(target ${srcfile} NAME_WE)
    add_wast_test(TARGET ${target}
            INCLUDE_FOLDERS ${STANDARD_INCLUDE_FOLDERS}
            LIBRARIES ${libbcwasm} ${libc++} ${libc}
            DESTINATION_FOLDER ${CMAKE_CURRENT_BINARY_DIR}
            )
endforeach()