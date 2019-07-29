//
// Created by zhou.yang on 2018/11/17.
//
#include "bcwasm/storagetype.hpp"
#include "../unittest.hpp"
#include "bcwasm/print.hpp"
#define ENABLE_TRACE

#define CAT()
#define STORAGE_NAME(TYPE) s##TYPE

#define SET_GET(TYPE, BASETYPE, STORAGE)\
    char STORAGE_NAME(TYPE)[] = #TYPE; \
    TEST_CASE(SetGet, TYPE) { \
        TRACE("TEST", #TYPE); \
        { \
            bcwasm::StorageType<STORAGE_NAME(TYPE), BASETYPE> v(STORAGE); \
        } \
        bcwasm::StorageType<STORAGE_NAME(TYPE), BASETYPE> v2(STORAGE); \
        ASSERT(v2 == STORAGE);\
    }

SET_GET(uint8_t, uint8_t, 1)
SET_GET(int8_t, int8_t, 1)
SET_GET(uint16_t, uint16_t, 1)
SET_GET(int16_t, int16_t, 1)
SET_GET(uint32_t, uint32_t, 1)
SET_GET(int32_t, int32_t, 1)
SET_GET(uint64_t, uint64_t, 1)
SET_GET(int64_t, int64_t, 1)
SET_GET(string, std::string, "hello")


UNITTEST_MAIN() {
    RUN_TEST(SetGet, uint8_t)
    RUN_TEST(SetGet, int8_t)
    RUN_TEST(SetGet, uint16_t)
    RUN_TEST(SetGet, int16_t)
    RUN_TEST(SetGet, uint32_t)
    RUN_TEST(SetGet, int32_t)
    RUN_TEST(SetGet, uint64_t)
    RUN_TEST(SetGet, int64_t)
    RUN_TEST(SetGet, string)
}
