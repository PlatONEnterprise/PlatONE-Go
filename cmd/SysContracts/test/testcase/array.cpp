//
// Created by zhou.yang on 2018/11/20.
//

#define ENABLE_TRACE
#include "bcwasm/db/array.hpp"
#include "../unittest.hpp"

char arrayStrName[] = "arraystr";
char arrayIntName[] = "arrayint";
char arraySetName[] = "arrayset";

typedef bcwasm::db::Array <arrayIntName, int, 20> ArrayInt;
typedef bcwasm::db::Array <arrayStrName, std::string, 2> ArrayStr;
typedef bcwasm::db::Array <arraySetName, std::string, 10> ArraySet;

TEST_CASE(array, batch) {
    {
        ArrayInt arrayInt;
        for (size_t i = 0; i < 20; i++) {
            arrayInt[i] = i;
        }
        for (size_t i = 0; i < 10; i++) {
            arrayInt[i] = 0;
        }
    }

    {
        ArrayInt arrayInt;
        for (size_t i = 0; i < 10; i++) {
            ASSERT(arrayInt[i] == 0, "array[", i, "]", arrayInt[i]);
        }

        for (size_t i = 10; i < 20; i++) {
            ASSERT(arrayInt[i] == i, "array[", i, "]", arrayInt[i]);
        }
    }

    {
        DEBUG("test iterator");
        ArrayInt arrayInt;
        ArrayInt::Iterator iter = arrayInt.begin();
        for (size_t i = 0; i < 10 && iter != arrayInt.end(); i++, iter++) {
            TRACE("i:", i, "iter:", *iter);
            ASSERT(*iter == 0);
        }

        for (size_t i = 10; i < 20 && iter != arrayInt.end(); i++, iter++) {
            TRACE("i:", i, "iter:", *iter);
            ASSERT(*iter == i, "iter:", *iter, "i:", i);
        }

        ArrayInt::ConstIterator citer = arrayInt.cbegin();
        for (size_t i = 0; i < 10 && citer != arrayInt.cend(); i++, citer++) {
            TRACE("i:", i, "iter:", *citer);
            ASSERT(*citer == 0);
        }

        for (size_t i = 10; i < 20 && citer != arrayInt.cend(); i++, citer++) {
            TRACE("i:", i, "iter:", *citer);
            ASSERT(*citer == i, "iter:", *citer, "i:", i);
        }
    }

    {
        DEBUG("test  reserve iterator");
        ArrayInt arrayInt;

        ArrayInt::ConstIterator citer = arrayInt.cbegin();
        for (size_t i = 0; i < 10 && citer != arrayInt.cend(); i++, citer++) {
            ASSERT(*citer == 0);
        }

        for (size_t i = 10; i < 20 && citer != arrayInt.cend(); i++, citer++) {
            ASSERT(*citer == i, "iter:", *citer, "i:", i);
        }

        ArrayInt::ReverseIterator iter = arrayInt.rbegin();
        for (size_t i = 19; i >= 10 && iter != arrayInt.rend(); i--, iter++) {
            ASSERT(*iter == i, "iter:", *iter, "i:", i);
        }

        for (size_t i = 9; i >= 0 && iter != arrayInt.rend(); i--, iter++) {
            ASSERT(*iter == 0, "iter:", *iter);
        }

        ArrayInt::ConstReverseIterator criter = arrayInt.crbegin();
        for (size_t i = 19; i >= 10 && criter != arrayInt.crend(); i--, ++criter) {
            ASSERT(*criter == i, "criter:", *criter, "i:", i);
        }

        for (size_t i = 9; i >= 0 && criter != arrayInt.crend(); i--, criter++) {
            ASSERT(*criter == 0, "criter:", *criter, "i:", i);
        }
    }

}

TEST_CASE(array, open){
    {
        ArrayStr arrayStr;
        arrayStr[0] = "hello";
        arrayStr[1] = "world";
    }

    {
        DEBUG("test reopen");
        ArrayStr arrayStr;

        ASSERT(arrayStr[0] == "hello", "arrayStr[0]:", arrayStr[0]);
        ASSERT(arrayStr[1] == "world", "arrayStr[1]:", arrayStr[1]);
    }

    {
        DEBUG("test del");
        ArrayStr arrayStr;
        arrayStr[0] = "";
    }

    {
        DEBUG("test reopen deled");
        ArrayStr arrayStr;
        ASSERT(arrayStr[0] == "", "arrayStr[0]:", arrayStr[0]);
    }

}

TEST_CASE(array, set) {
    ArraySet array;
    array[0] = "hello";
    array.setConst(0, "helloworld");
    ASSERT_EQ(array[0], "helloworld");
    array.setConst(1, "world");
    ASSERT_EQ(array[1], "world");
}




UNITTEST_MAIN() {
    RUN_TEST(array, batch)
    RUN_TEST(array, open)
    RUN_TEST(array, set);
}
