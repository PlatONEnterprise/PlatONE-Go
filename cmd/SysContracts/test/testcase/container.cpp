//
// Created by zhou.yang on 2018/11/21.
//

//
// Created by zhou.yang on 2018/11/20.
//

#define ENABLE_TRACE
#include "bcwasm/storagetype.hpp"
#include "../unittest.hpp"

char vectorStrName[] = "vectorstr";



TEST_CASE(container, vector){
    {
        bcwasm::Vector<vectorStrName, std::string> vectorStr;
        vectorStr->push_back("hello");
        vectorStr->push_back("world");
    }

    {
        DEBUG("test reopen");
        bcwasm::Vector<vectorStrName, std::string> vectorStr;
        ((*vectorStr)[0] == "hello", "");
        ASSERT((*vectorStr)[1] == "world", "(*vectorStr)[1]:", (*vectorStr)[1]);
    }

    {
        DEBUG("test del");
        bcwasm::Vector<vectorStrName, std::string> vectorStr;
        vectorStr->clear();
        ASSERT(vectorStr->size() == 0, "size:", vectorStr->size());
    }

    {
        DEBUG("test reopen deled");
        bcwasm::Vector<vectorStrName, std::string> vectorStr;
        ASSERT(vectorStr->size() == 0, "size:", vectorStr->size());
    }

}

char arrayStrName[] = "arraystr";
TEST_CASE(container, array){
    {
        bcwasm::Array<arrayStrName, std::string, 2> arrayStr;
        (*arrayStr)[0] = "hello";
        (*arrayStr)[1] = "world";
    }

    {
        DEBUG("test reopen");
        bcwasm::Array<arrayStrName, std::string, 2> arrayStr;
        ASSERT((*arrayStr)[0] == "hello", "(*arrayStr)[0]:", (*arrayStr)[0]);
        ASSERT((*arrayStr)[1] == "world", "(*arrayStr)[1]:", (*arrayStr)[1]);
    }

}

char setStrName[] = "setName";
TEST_CASE(container, set){
    {
        bcwasm::Set<setStrName, std::string> set;
        set->insert("hello");
        set->insert("world");
    }

    {
        DEBUG("test reopen");
        bcwasm::Set<setStrName, std::string> set;
        ASSERT(set->find("hello") != set->end());
        ASSERT(set->find("world") != set->end());
    }
}

char mapStrName[] = "mapName";
TEST_CASE(container, map){
    {
        bcwasm::Map<mapStrName, std::string, std::string> map;
        map->insert(std::pair<std::string, std::string>("hello", "world"));
    }

    {
        DEBUG("test reopen");
        bcwasm::Map<mapStrName, std::string, std::string> map;
        ASSERT(map->find("hello") != map->end());
        ASSERT((*map)["hello"] == "world");
    }
}

char tupleStrName[] = "tupleName";
TEST_CASE(container, tuple){
    {
        bcwasm::Tuple<tupleStrName, std::string, std::string> tuple;
        (*tuple) = std::make_tuple("hello", "world");
    }

    {
        DEBUG("test reopen");
        bcwasm::Tuple<tupleStrName, std::string, std::string> tuple;
        ASSERT(std::get<0>(*tuple) == "hello");
        ASSERT(std::get<1>(*tuple) == "world");
    }
}



UNITTEST_MAIN() {
    RUN_TEST(container, vector);
    RUN_TEST(container, array);
    RUN_TEST(container, set);
    RUN_TEST(container, map);
    RUN_TEST(container, tuple);
}






