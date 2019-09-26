#include "../unittest.hpp"
#include "bcwasm/common.h"

#include <string>

using namespace bcwasm;

extern "C" {
    // dst = src + 1
    void bigintAdd(const byte* src, size_t src_len, byte* dst, size_t& dst_len);
}

// test bcwasm::bigint
TEST_CASE(bigint, bigintAdd) {
    bigint i = std::numeric_limits<__int128>::max();
    bytes src(512/8);
    toBigEndian(i, src);
    bytes dst(512/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);

    bigint ii = fromBigEndian<bigint, bytes>(dst);

    ASSERT(ii == (i+1));
}

TEST_CASE(bigint, u64Add) {
    u64 i = std::numeric_limits<uint32_t>::max();
    bytes src(64/8);
    toBigEndian(i, src);
    bytes dst(64/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);

    u64 ii = fromBigEndian<u64, bytes>(dst);

    ASSERT(ii == (i+1));
}

TEST_CASE(bigint, u128Add) {
    u128 i = std::numeric_limits<__int128>::max();
    bytes src(128/8);
    toBigEndian(i, src);
    bytes dst(128/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);

    u128 ii = fromBigEndian<u128, bytes>(dst);

    ASSERT(ii == i+1);
}

TEST_CASE(bigint, u160Add) {
    u160 i = std::numeric_limits<unsigned __int128>::max();
    bytes src(160/8);
    toBigEndian(i, src);
    bytes dst(160/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);
    
    u160 ii = fromBigEndian<u160, bytes>(dst);
    ASSERT(ii == i+1);
}

TEST_CASE(bigint, u256Add) {
    u256 i = std::numeric_limits<unsigned __int128>::max();
    bytes src(256/8);
    toBigEndian(i, src);
    bytes dst(256/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);

    u256 ii = fromBigEndian<u256, bytes>(dst);
    ASSERT(ii == i+1);
}

TEST_CASE(bigint, u512Add) {
    u512 i = std::numeric_limits<u256>::max();
    bytes src(512/8);
    toBigEndian(i, src);
    bytes dst(512/8);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);

    u512 ii = fromBigEndian<u512, bytes>(dst);
    ASSERT(ii == i + 1);
}

TEST_CASE(bigint, u64Overflow) {
    u64 i = std::numeric_limits<uint64_t>::max();
    bytes src(8);
    toBigEndian(i, src);
    bytes dst(16);
    size_t dst_len;
    bigintAdd(src.data(), src.size(), dst.data(), dst_len);
    dst.resize(dst_len);
    DEBUG("dst_len:", dst_len);
    
    u64 ii = fromBigEndian<u64, bytes>(dst);

    ASSERT(ii == (i+1));
}

TEST_CASE(bigint, convertStr) {
    u64 i = 1234567890;
    ASSERT(i.convert_to<std::string>() == "1234567890");

    u128 i128 = 1234567890;
    ASSERT(i128.convert_to<std::string>() == "1234567890");

    u160 i160 = 1234567890;
    ASSERT(i160.convert_to<std::string>() == "1234567890");

    u256 i256 = 1234567890;
    ASSERT(i256.convert_to<std::string>() == "1234567890");

    u512 i512 = 1234567890;
    ASSERT(i512.convert_to<std::string>() == "1234567890");

    std::string hex = "0x00ff";
    bytes bs = fromHex(hex);
    u64 ii = fromBigEndian<u64, bytes>(bs);
    ASSERT(ii == 255);

    bs.resize(8);
    toBigEndian(ii, bs);
    hex = toHex(bs.begin(), bs.end(), "0x");
    ASSERT(hex == "0x00000000000000ff");
}

TEST_CASE(bigint, operator) {
    u256 i = 100;
    ASSERT(i-50 == 50);
    ASSERT(i+10 == 110);
    ASSERT(i/2 == 50);
    ASSERT(i*2 == 200);
}

UNITTEST_MAIN() {
    RUN_TEST(bigint, bigintAdd);
    RUN_TEST(bigint, u64Add);
    RUN_TEST(bigint, u128Add);
    RUN_TEST(bigint, u160Add);
    RUN_TEST(bigint, u256Add);
    RUN_TEST(bigint, u512Add);
    RUN_TEST(bigint, u64Overflow);
    RUN_TEST(bigint, convertStr);
    RUN_TEST(bigint, operator);
}
