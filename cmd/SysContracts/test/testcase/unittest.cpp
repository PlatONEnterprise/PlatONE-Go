//
// Created by zhou.yang on 2018/11/21.
//

#include "../unittest.hpp"

TEST_CASE(hello, world) {
    ASSERT(true, "hello");
    int i = 1;
    int j = 2;
    ASSERT_NE(i, j, "xxxx");
}


UNITTEST_MAIN() {
    RUN_TEST(hello, world)
}