//
// Created by zhou.yang on 2018/10/20.
//
#pragma once

#include <stdlib.h>
#include "print.hpp"

namespace bcwasm {
    /**
     * @brief Throw an exception after printing the message
     * 
     * @tparam Arg Starting parameter type
     * @tparam Args Variable parameter type
     * @param a Starting parameter
     * @param args Variable parameter
     */
    template<typename Arg, typename... Args>
    void bcwasmThrow(Arg&& a, Args&&... args){
        println(a, args...);
        abort();
    }
}