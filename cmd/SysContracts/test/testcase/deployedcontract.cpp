//
// Created by zhou.yang on 2018/11/27.
//

#include "../unittest.hpp"
#include "../log.h"
#include "bcwasm/fixedhash.hpp"
#include "bcwasm/deployedcontract.hpp"
using namespace bcwasm;


class LOG_TEST {
public:
    ~LOG_TEST(){
        ::clearLog();
    }
};


TEST_CASE(test, contract){

    DeployedContract contract(Address("0xa0b21d5bcc6af4dda0579174941160b9eecb6918"));
    {
        LOG_TEST logTest;
        contract.callString("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }

    {
        LOG_TEST logTest;
        contract.delegateCallString("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }

    {
        LOG_TEST logTest;
        contract.callInt64("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }

    {
        LOG_TEST logTest;
        contract.delegateCallInt64("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }

    {
        LOG_TEST logTest;
        contract.call("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }

    {
        LOG_TEST logTest;
        contract.delegateCall("hello", 123, "1234");
        ASSERT(test::getLog() == "a0b21d5bcc6af4dda0579174941160b9eecb6918 d98800000000000000098568656c6c6f840000007b8431323334")
    }
}

UNITTEST_MAIN() {
    RUN_TEST(test, contract)
};
