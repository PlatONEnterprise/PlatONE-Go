//
// Created by zhou.yang on 2018/11/26.
//

#include "../unittest.hpp"
#include "../state.h"
#include "bcwasm/state.hpp"



TEST_CASE(test, state) {
    std::string stateJson = R"E({
	"gasPrice":123,
 	 "blockHash": {
		"1":"0xfbf4a7f56654ef98af0f908ac8d46a14ac3736424b31fea297d93281c21fc2ed",
		"2":"0xfbf4a7f56654ef98af0f908ac8d46a14ac3736424b31fea297d93281c21fc2de"
		},
     "number":234,
     "gasLimit" : 1234,
	 "timestamp" : 1543284111,
     "coinbase" : "0xa0b21d5bcc6af4dda0579174941160b9eecb6917",
     "balance" : 1234,
     "origin" : "0xa0b21d5bcc6af4dda0579174941160b9eecb6918",
     "caller" : "0xa0b21d5bcc6af4dda0579174941160b9eecb6919",
     "value" : 123,
     "address" : "0xa0b21d5bcc6af4dda0579174941160b9eecb6920",
     "nonce" : 1234,
	 "account" : {
		"0xa0b21d5bcc6af4dda0579174941160b9eecb6918":200,
		"0xa0b21d5bcc6af4dda0579174941160b9eecb6920":1234
	 }
    })E";
    ::setStateDB(stateJson.data(), stateJson.length());
    ASSERT(::number() == 234);
    ASSERT(::gasPrice() == 123);
    ASSERT(bcwasm::blockHash(2).toString() == "fbf4a7f56654ef98af0f908ac8d46a14ac3736424b31fea297d93281c21fc2de");
    ASSERT(::gasLimit() == 1234);
    ASSERT(::timestamp() == 1543284111);
    ASSERT(bcwasm::coinbase().toString() == "a0b21d5bcc6af4dda0579174941160b9eecb6917");
    ASSERT(bcwasm::balance() == 1234, bcwasm::balance().convert_to<std::string>().c_str());
    ASSERT(bcwasm::origin().toString() == "a0b21d5bcc6af4dda0579174941160b9eecb6918");
    ASSERT(bcwasm::caller().toString() == "a0b21d5bcc6af4dda0579174941160b9eecb6919");
    ASSERT(bcwasm::callValue() == 123);
    ASSERT(bcwasm::address().toString() == "a0b21d5bcc6af4dda0579174941160b9eecb6920");
    ASSERT(::getCallerNonce() == 1234);
    bcwasm::Address to("a0b21d5bcc6af4dda0579174941160b9eecb6918");
    bcwasm::u256 amount = 200;
    ASSERT(bcwasm::callTransfer(to, amount) == 0);
    ASSERT(bcwasm::balance() == 1034, bcwasm::balance().convert_to<std::string>());
}

UNITTEST_MAIN() {
    RUN_TEST(test, state)
}
