//
// Created by zhou.yang on 2018/11/20.
//

#define ENABLE_TRACE
#include "bcwasm/db/map.hpp"
#include "../unittest.hpp"

char mapStrName[] = "mapstr";

typedef bcwasm::db::Map<mapStrName, std::string, std::string> MapStr;

char mapInsertName[] = "mapinsert";

typedef bcwasm::db::Map<mapInsertName, std::string, std::string> MapInsert;

TEST_CASE(map, operator) {
    {
        MapStr map;
        map.insert("hello", "world");
        ASSERT(map["hello"] == "world", map["hello"].c_str());
        TRACE("map.size:", map.size());
    }

    {
        MapStr map;
        TRACE("map.size:", map.size());
        ASSERT(map["hello"] == "world", map["hello"].c_str());
    }

    {
        DEBUG("test insert");
        MapStr map;
        TRACE("map.size:", map.size());
        map["hello1"] = "world";
        ASSERT(map.size() == 2);
    }

    {
        DEBUG("test del");
        MapStr map;
        map.del("hello");
        ASSERT(map.size() == 1);
    }

    {
        DEBUG("test reopen del");
        MapStr map;
        ASSERT(map.size() == 1);
    }

    {
        DEBUG("test iterator");
        MapStr map;
        size_t size = map.size();
        TRACE("map size:", size);
        size_t count = 0;
        for (MapStr::Iterator iter = map.begin(); iter != map.end(); iter++) {
            TRACE("iter second:", iter->second());
            ASSERT(iter->second() == "world");
            count++;
        }
        TRACE("count:", count, "size:", size);
        ASSERT(count == size);
        count = 0;
        for (MapStr::ReverseIterator iter = map.rbegin(); iter != map.rend(); ++iter) {
            TRACE("iter....");
            TRACE("iter second:", iter->second());
            ASSERT(iter->second() == "world", "reverse iterator error");
            TRACE("iter second:", iter->second());
            count++;
        }
        TRACE("count:", count, "size:", size);
        ASSERT(count == size);

        count = 0;
        for (MapStr::ConstIterator citer = map.cbegin(); citer != map.cend(); citer++) {
            TRACE("iter second:", citer->second());
            ASSERT(citer->second() == "world");
            count++;
        }
        TRACE("count :", count, "size:", size);
        ASSERT(count == size);
        count = 0;
        for (MapStr::ConstReverseIterator criter = map.crbegin(); criter != map.crend(); ++criter) {
            TRACE("iter....");
            TRACE("iter second:", criter->second());
            ASSERT(criter->second() == "world", "reverse iterator error");
            TRACE("iter second:", criter->second());
            count++;
        }
        TRACE("count:", count, "size:", size);
        ASSERT(count == size);
    }


}


TEST_CASE(map, insert) {
    MapInsert map;
    map["hello"] = "world";
    map.insertConst("hello", "helloworld");
    ASSERT(map["hello"] == "helloworld");
}



UNITTEST_MAIN() {
    RUN_TEST(map, operator);
    RUN_TEST(map, insert);
}