add_wast_executable(TARGET roleRegister 
     #Multiple cpp files use SOURCE_FILES, and the name of TARGET is the same as the cpp file name with BCWASM_ABI
     #SOURCE_FILES multi filenames 
     INCLUDE_FOLDERS ${STANDARD_INCLUDE_FOLDERS} 
     LIBRARIES  ${libbcwasm} ${libc++} ${libc}
     DESTINATION_FOLDER ${CMAKE_CURRENT_BINARY_DIR})
