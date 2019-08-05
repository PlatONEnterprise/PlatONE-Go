#include "../unittest.hpp"
#include "bcwasm/datastream.h"

#define SINGLE(...) __VA_ARGS__
#define DATASTREAM_CASE(NAME, TYPE, DATA) \
    TEST_CASE(DataStream, NAME)  { \
                TYPE v = DATA; \
                char buf[128]; \
                bcwasm::DataStream<char*> ds(buf, 128); \
                ds << v; \
                TYPE v2; \
                ds.seekp(0); \
                ds >> v2; \
                ASSERT(v == v2); \
    }


struct Pair {
    int a;
    double d;
    bool operator==(const Pair &p) const { return a == p.a && std::abs(d - p.d) < 1e-20;}
};


struct StaticArray {
    int a[2];
    bool operator==(const StaticArray &o) const { return a[0] == o.a[0] && a[1] == o.a[1]; }
};



DATASTREAM_CASE(bool_true, bool, true)
DATASTREAM_CASE(bool_false, bool, false)
DATASTREAM_CASE(int8_t, int8_t, -123)
DATASTREAM_CASE(uint8_t, uint8_t, 127)
DATASTREAM_CASE(int16_t, int16_t, -1234)
DATASTREAM_CASE(uint16_t, uint16_t, 12345)
DATASTREAM_CASE(int32_t, int32_t, -12345566)
DATASTREAM_CASE(uint32_t, uint32_t, 12345676)
DATASTREAM_CASE(int64_t, int64_t, -1234342432444)
DATASTREAM_CASE(uint64_t, uint64_t, 1234342432444)
DATASTREAM_CASE(float, float, 1.234f)
DATASTREAM_CASE(double, double, 0.234334444)
DATASTREAM_CASE(string, std::string, "hello")
DATASTREAM_CASE(vector, SINGLE(std::vector<int>), SINGLE({10,20,30}))
DATASTREAM_CASE(vector_empty, SINGLE(std::vector<int>), SINGLE({}))
DATASTREAM_CASE(array, SINGLE(std::array<int,3>), SINGLE({{10,20,30}}))
DATASTREAM_CASE(map, SINGLE(std::map<int, std::string>), SINGLE({{1, "apple"}, {2, "cat"}, {3, "panda"}}))
DATASTREAM_CASE(tuple, SINGLE(std::tuple<std::string, std::string, double>), SINGLE({"1", "ghs", 5.341}))
DATASTREAM_CASE(StaticArray, StaticArray, SINGLE({{10,20}}))
DATASTREAM_CASE(Pair, Pair, SINGLE({1, 1.23456}));
DATASTREAM_CASE(FixedHash, SINGLE(bcwasm::FixedHash<5> ), SINGLE({1,2,3,4,5}))
DATASTREAM_CASE(u256, bcwasm::u256, 12312312312312312)


UNITTEST_MAIN(){
    RUN_TEST(DataStream, bool_true)
    RUN_TEST(DataStream, bool_false)
    RUN_TEST(DataStream, int8_t)
    RUN_TEST(DataStream, uint8_t)
    RUN_TEST(DataStream, int16_t)
    RUN_TEST(DataStream, uint16_t)
    RUN_TEST(DataStream, int32_t)
    RUN_TEST(DataStream, uint32_t)
    RUN_TEST(DataStream, int64_t)
    RUN_TEST(DataStream, uint64_t)
    RUN_TEST(DataStream, float)
    RUN_TEST(DataStream, double)
    RUN_TEST(DataStream, string)
    RUN_TEST(DataStream, vector)
    RUN_TEST(DataStream, vector_empty)
    RUN_TEST(DataStream, array)
    RUN_TEST(DataStream, map)
    RUN_TEST(DataStream, tuple)
    RUN_TEST(DataStream, StaticArray)
    RUN_TEST(DataStream, Pair)
    RUN_TEST(DataStream, FixedHash)
    RUN_TEST(DataStream, u256)
}

