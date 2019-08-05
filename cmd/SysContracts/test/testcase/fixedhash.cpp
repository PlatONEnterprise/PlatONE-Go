//
// Created by zhou.yang on 2018/11/22.
//

#include "../unittest.hpp"
#include "bcwasm/fixedhash.hpp"
#include "bcwasm/state.hpp"
using namespace bcwasm;
TEST_CASE(fixedhash, Compare) {
    DEBUG("sfdasfadfa");
    FixedHash<4> h1(::bcwasm::sha3("abcd"));
    FixedHash<4> h2(::bcwasm::sha3("abcd"));
    FixedHash<4> h3(::bcwasm::sha3("aadd"));
    FixedHash<4> h4(0xBAADF00D);
    FixedHash<4> h5(0xAAAAAAAA);
    FixedHash<4> h6(0xBAADF00D);

    ASSERT_EQ(h1, h2);
    ASSERT_NE(h2, h3);

    ASSERT(h4 > h5);
    ASSERT(h5 < h4);
    ASSERT(h6 <= h4);
    ASSERT(h6>= h4);
}

TEST_CASE(FixedHash, XOR)
{
    FixedHash<2> h1("0xAAAA");
    FixedHash<2> h2("0xBBBB");

    ASSERT_EQ((h1 ^ h2), FixedHash<2>("0x1111"));
    h1 ^= h2;
    ASSERT_EQ(h1, FixedHash<2>("0x1111"));
}

TEST_CASE(FixedHash, OR)
{
    FixedHash<4> h1("0xD3ADB33F");
    FixedHash<4> h2("0xBAADF00D");
    FixedHash<4> res("0xFBADF33F");

    ASSERT_EQ((h1 | h2), res);
    h1 |= h2;
    ASSERT_EQ(h1, res);
}

TEST_CASE(FixedHash, AND)
{
    FixedHash<4> h1("0xD3ADB33F");
    FixedHash<4> h2("0xBAADF00D");
    FixedHash<4> h3("0x92aDB00D");

    ASSERT_EQ((h1 & h2), h3);
    h1 &= h2;
    ASSERT_EQ(h1, h3);
}

TEST_CASE(FixedHash, invert)
{
    FixedHash<4> h1("0xD3ADB33F");
    FixedHash<4> h2("0x2C524CC0");

    ASSERT_EQ(~h1, h2);
}

TEST_CASE(FixedHash, contains)
{
    FixedHash<4> h1("0xD3ADB331");
    FixedHash<4> h2("0x0000B331");
    FixedHash<4> h3("0x0000000C");

    ASSERT(h1.contains(h2));
    ASSERT(!h1.contains(h3));
}

UNITTEST_MAIN() {
    RUN_TEST(fixedhash, Compare)
    RUN_TEST(FixedHash, XOR)
    RUN_TEST(FixedHash, OR)
    RUN_TEST(FixedHash, AND)
    RUN_TEST(FixedHash, invert)
    RUN_TEST(FixedHash, contains)
}