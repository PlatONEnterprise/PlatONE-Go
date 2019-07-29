file(GLOB CRYPT_SOURCES  "musl/src/crypt/*.c")
file(GLOB CTYPE_SOURCES  "musl/src/ctype/*.c")
file(GLOB ENV_SOURCES  "musl/src/env/*.c")
file(GLOB ERRNO_SOURCES  "musl/src/errno/*.c")
file(GLOB EXIT_SOURCES  "musl/src/exit/*.c")
file(GLOB LOCALE_SOURCES "musl/src/locale/*.c")
file(GLOB MATH_SOURCES  "musl/src/math/*.c")
file(GLOB MBYTE_SOURCES  "musl/src/multibyte/*.c")
file(GLOB MISC_SOURCES "musl/src/misc/*.c")
file(GLOB SEARCH_SOURCES "musl/src/search/*.c")
file(GLOB STDIO_SOURCES "musl/src/stdio/*.c")
file(GLOB STDLIB_SOURCES "musl/src/stdlib/*.c")
file(GLOB STRING_SOURCES "musl/src/string/*.c")
file(GLOB TIME_SOURCES "musl/src/time/*.c")
file(GLOB THREAD_SOURCES "musl/src/thread/*.c") #only for __lock __unlock
set(INTERNAL_SOURCES musl/src/internal/floatscan.c musl/src/internal/intscan.c musl/src/internal/shgetc.c musl/src/internal/libc.c)

add_wast_library(TARGET libc
  NOWARNINGS
  SOURCE_FILES ${CRYPT_SOURCES} ${CTYPE_SOURCES} ${ENV_SOURCES} ${ERRNO_SOURCES} ${EXIT_SOURCES}  ${LOCALE_SOURCES} ${MATH_SOURCES}
  ${MBYTE_SOURCES} ${MISC_SOURCES} ${SEARCH_SOURCES} ${STDIO_SOURCES} ${STDLIB_SOURCES} ${STRING_SOURCES} ${TIME_SOURCES} ${THREAD_SOURCES} ${INTERNAL_SOURCES}
  INCLUDE_FOLDERS ${CMAKE_SOURCE_DIR}/libc/musl/include
                  ${CMAKE_SOURCE_DIR}/libc/musl/src/internal
                  ${CMAKE_SOURCE_DIR}/libc/musl/arch/bcwasm
                  ${CMAKE_SOURCE_DIR}/
  DESTINATION_FOLDER ${CMAKE_CURRENT_BINARY_DIR}
)
