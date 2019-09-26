//
// Created by zhou.yang on 2018/11/19.
//
#define ENABLE_TRACE
#include "bcwasm/db/list.hpp"
#include "../unittest.hpp"

char listStrName[] = "liststr";
char listIntName[] = "listint";
char listPushName[] = "listPush";
char listInsertName[] = "listInsert";

typedef bcwasm::db::List < listIntName, int > ListInt;
typedef bcwasm::db::List < listStrName, std::string > ListStr;
typedef bcwasm::db::List < listPushName, std::string > ListPush;
typedef bcwasm::db::List < listInsertName, std::string > ListInsert;

TEST_CASE(list, push) {
    {
        ListPush pl;
        for (int i = 0; i < 500; i++) {
            pl.push("hello");
        }

        for (int i = 0; i < 400; i++) {
            pl.del(0);
        }
    }
    {
        ListPush pl;
        for (int i = 0; i < 100; i++) {
           ASSERT_EQ(pl[i], "hello");
        }
    }
}
TEST_CASE(list, batch) {
    {
        ListInt listInt;
        for (size_t i = 0; i < 20; i++) {
            listInt.push(i);
        }
        for (size_t i = 0; i < 10; i++) {
            listInt.del(i);
        }
    }

    {
        ListInt listInt;
        for (size_t i = 0; i < 10; i++) {
            ASSERT(listInt[i] == i+10);
        }
    }

}
TEST_CASE(list, opendel){
        {
            ListStr listStr;
            listStr.push("hello");
            listStr.push("world");
        }

        {
            DEBUG("test reopen");
            ListStr listStr;
            ASSERT(listStr[0] == "hello");
            ASSERT(listStr[1] == "world");
            TRACE("listStr size:", listStr.size());
        }

        {
            DEBUG("test del");
            ListStr listStr;
            TRACE("listStr size:", listStr.size());
            listStr.del(0);
            TRACE("listStr size:", listStr.size());
            ASSERT(listStr.size() == 1);
        }

        {
            DEBUG("test reopen deled");
            ListStr listStr;
            ASSERT(listStr.size() == 1);
            ASSERT(listStr[0] == "world", "listStr[0] ", listStr[0]);
        }
        {
            DEBUG("test reopen iterator");
            ListStr listStr;
            listStr.push("world");
            ASSERT(listStr.size() == 2);

            size_t size = listStr.size();
            TRACE("list size:", size);
            size_t count = 0;
            for (ListStr::Iterator iter = listStr.begin(); iter != listStr.end(); iter++) {
                TRACE("++++++++++");
                ASSERT(*iter == "world");
                count++;
                TRACE("--------");
            }
            ASSERT(count == size);
            count = 0;
            for (ListStr::ReverseIterator iter = listStr.rbegin(); iter != listStr.rend(); iter++) {
                ASSERT(*iter == "world");
                count++;
            }
            ASSERT(count == size);

            count = 0;
            for (ListStr::ConstIterator iter = listStr.cbegin(); iter != listStr.cend(); iter++) {
                ASSERT(*iter == "world");
                count++;
            }

            ASSERT(count == size);
            count = 0;
            for (ListStr::ConstReverseIterator iter = listStr.crbegin(); iter != listStr.crend(); iter++) {
                ASSERT(*iter == "world");
                count++;
            }
            ASSERT(count == size);
        }

}

TEST_CASE(list, insert){
    ListInsert listInsert;
    listInsert.push("hello");
    listInsert.setConst(0,  "helloworld");
    ASSERT(listInsert[0] == "helloworld");
}



UNITTEST_MAIN() {
    RUN_TEST(list, push)
    RUN_TEST(list, batch)
    RUN_TEST(list, opendel)
    RUN_TEST(list, insert)
}
