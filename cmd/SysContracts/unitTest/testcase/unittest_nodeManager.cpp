//
// Created by zhou.yang on 2018/11/21.
//

#include "../unittest.hpp"
#include "../../systemContract/nodeManager/nodeManager.cpp"
#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include "bcwasm/print.h"
using namespace systemContract;

TEST_CASE(test, NodeManager){
    NodeManager nm;
    nm.add("{\"name\":\"node1\",\"owner\":\"0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC\",\"desc\":\"i love this world\",\"type\":1,\"publicKey\":\"acb2281452fb9fc25d40113fb6afe82b498361de0ee4ce69f55c180bb2afce2c5a00f97bfbe0270fadba174264cdf6da76ba334a6380c0005a84e8a6449c2502\",\"externalIP\":\"127.0.0.1\",\"internalIP\":\"127.0.0.1\",\"rpcPort\":4789,\"p2pPort\":14789,\"status\":1,\"root\":true}");
    bcwasm::println("node list:", nm.getAllNodes());
}

UNITTEST_MAIN() {
    RUN_TEST(test, NodeManager)
}