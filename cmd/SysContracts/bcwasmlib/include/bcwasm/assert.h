//
// Created by zhou.yang on 2018/11/1.
//

#pragma once

#include "exception.h"

namespace bcwasm{
    /**
     * @brief BCWasm assertion.
     *
     */
    #define BCWasmAssert(A, ...) ::bcwasm::assertAux(A, #A, __LINE__, __FILE__, __func__, ##__VA_ARGS__)
    /**
     * @brief Assertion A equals B.
     *
     */
    #define BCWasmAssertEQ(A, B, ...) BCWasmAssert(((A)==(B)),##__VA_ARGS__)
    /**
     * @brief Assertion A not equal to B.
     */
    #define BCWasmAssertNE(A, B, ...) BCWasmAssert(((A)!=(B)), ##__VA_ARGS__)

    /**
     * @brief BCWasm assertion implementation, assertion failure output failure location and error message.
     * @param cond      Assertion condition.
     * @param conndStr  Assertion prompt.
     * @param line      The line number of the code that failed the assertion.
     * @param file      Code file with assertion failure.
     * @param func      Function with assertion failure.
     * @param args      Argument list.
     */
    template<typename... Args>
    inline void assertAux(bool cond, const char *condStr, int line, const char *file, const char *func, Args&&... args) {
        if (!cond) {
           bcwasmThrow("Assertion failed:", condStr, "func:", func, "line:", line, "file:", file, args...);
        }
    }
}

