//
// Created by zhou.yang on 2018/11/22.
//

#include "../unittest.hpp"
#include "bcwasm/RLP.h"
#include "bcwasm/fixedhash.hpp"

using namespace bcwasm;
TEST_CASE(rlp, int) {
    std::string data = "c3010203";
    RLPStream stream(3);
    stream << 1 << 2 << 3;
    std::string result = toHex(stream.out());
    ASSERT_EQ(result, data, result, data);
}

TEST_CASE(rlp, address) {
    std::string data = "9443355c787c50b647c425f594b441d4bd751951c1";
    RLPStream stream;
    stream << Address("0x43355c787c50b647c425f594b441d4bd751951c1", true);
    std::string result = toHex(stream.out());
    ASSERT_EQ(result, data, result, data);
}

TEST_CASE(rlp, array) {
    std::string data = "aa307834333335356337383763353062363437633432356635393462343431643462643735313935316331";
    RLPStream stream;
    stream << "0x43355c787c50b647c425f594b441d4bd751951c1";
    std::string result = toHex(stream.out());
    ASSERT_EQ(result, data, result, data);
}


UNITTEST_MAIN(){
    RUN_TEST(rlp, int)
    RUN_TEST(rlp, address)
    RUN_TEST(rlp, array)
}