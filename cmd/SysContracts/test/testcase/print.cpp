//
// Created by zhou.yang on 2018/11/26.
//

#include "../unittest.hpp"
#include "../log.h"
#include "bcwasm/print.h"
#include "bcwasm/fixedhash.hpp"

#include <limits>

using namespace bcwasm;

class LOG_TEST {
public:
    ~LOG_TEST(){
        ::clearLog();
    }
};

TEST_CASE(test, print){
    std::string l = "hello";
    {
        LOG_TEST logTest;
        prints(l.c_str());
        ASSERT(test::getLog() == "hello");
    }

    {
        LOG_TEST logTest;
        prints_l(l.c_str(), l.length());
        ASSERT(test::getLog() == "hello");
    }

    {
        LOG_TEST logTest;
        int64_t i = -99999999998;
        printi(i);
        ASSERT(test::getLog() == "-99999999998");
    }

    {
        LOG_TEST logTest;
        uint64_t i = 99999999998;
        printui(i);
        ASSERT(test::getLog() == "99999999998");
    }

    // test printi128
    {
        LOG_TEST logTest;
        int128_t i = 1;
        printi128(&i);
        ASSERT(test::getLog() == "1");
    }

    {
        LOG_TEST logTest;
        int128_t i = 0;
        printi128(&i);
        ASSERT(test::getLog() == "0");
    }

    {
        LOG_TEST logTest;
        __int128 i = std::numeric_limits<__int128>::lowest();
        printi128(&i);
        ASSERT(test::getLog() == "-170141183460469231731687303715884105728");
    }

    {
        LOG_TEST logTest;
        __int128 i = -1234567890;
        printi128(&i);
        ASSERT(test::getLog() == "-1234567890");
    }

    // test printui128
    {
        LOG_TEST logTest;
        unsigned __int128 i = std::numeric_limits<unsigned __int128>::max();
        printui128(&i);
        ASSERT(test::getLog() == "340282366920938463463374607431768211455");
    }

    {
        LOG_TEST logTest;
        unsigned __int128 i = 0;
        printui128(&i);
        ASSERT(test::getLog() == "0");
    }

    {
        LOG_TEST logTest;
        unsigned __int128 i = 1234567890;
        printui128(&i);
        ASSERT(test::getLog() == "1234567890");
    }

    // test print u256
    {
        LOG_TEST logTest;
        u256 i = std::numeric_limits<unsigned __int128>::max();
        print(i);
        ASSERT(test::getLog() == "340282366920938463463374607431768211455");
    }

    // test printsf
    {
        LOG_TEST logTest;
        float f = 1.0f/2.0f;
        printsf(f);
        ASSERT(test::getLog() == "0.5");

        logTest = LOG_TEST();
        f = 5.0f * -0.75f;
        printsf(f);
        ASSERT(test::getLog() == "-3.75");

        logTest = LOG_TEST();
        f = 2e-6f / 3.0f;
        printsf(f);
        ASSERT(test::getLog() == "6.6666666e-07");
    }

    // test printdf
    {
        LOG_TEST logTest;
        double f = 1.0 / 2.0;
        printdf(f);
        ASSERT(test::getLog() == "0.5");

        logTest = LOG_TEST();
        f = 5.0 * -0.75;
        printdf(f);
        ASSERT(test::getLog() == "-3.75");

        logTest = LOG_TEST();
        f = 2e-6 / 3.0;
        printdf(f);
        ASSERT(test::getLog() == "6.666666666666666e-07");
    }

    // test printqf
    {
        LOG_TEST logTest;
        long double f = 1.0l / 2.0l;
        printqf(&f);
        ASSERT(test::getLog() == "0.5");

        logTest = LOG_TEST();
        f = 5.0l * -0.75l;
        printqf(&f);
        ASSERT(test::getLog() == "-3.75");

        logTest = LOG_TEST();
        f = 2e-6l / 3.0l;
        printqf(&f);
        ASSERT(test::getLog() == "6.66666666666666667e-07");
    }


    {
        LOG_TEST logTest;
        FixedHash<4> f("d3adb33f");
        printhex(f.data(), f.size());
        ASSERT(test::getLog() == "d3adb33f");
    }

}

UNITTEST_MAIN() {
    RUN_TEST(test, print)
};

