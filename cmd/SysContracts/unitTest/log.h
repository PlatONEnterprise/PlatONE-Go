//
// Created by zhou.yang on 2018/11/26.
//

#pragma once

#include "bcwasm/assert.h"

extern "C" {
    size_t getTestLog(char *log, size_t size);
    size_t getTestLogSize();
    void clearLog();
}
namespace bcwasm {
    namespace test {
        std::string getLog() {
            std::vector<char> log;
            size_t size = ::getTestLogSize();
            log.resize(size);
            BCWasmAssert(::getTestLog((char*)log.data(), size) == size);
            return std::string(log.data(), size);
        }
    }
}


