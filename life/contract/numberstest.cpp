//auto create contract
#include "stdlib.h"
#include "string.h"
#include "string"
#include "bcwasm/bcwasm.hpp"

namespace demo { 
class Demo : public bcwasm::Contract{
    public:
    Demo(){}

    void init()
    {
        bcwasm::println("init success...");
    }
    int128_t addLong (int128_t i1,int128_t i2) const
    {
        return i1 + i2;
    }
    uint128_t addUlong (uint128_t i1,uint128_t i2) const
    {
        return i1 + i2;
    }
    float  addFloat(float f1 ,float f2) const
    {
        return f1+f2;
    }
    double addDouble(double f1 ,double f2) const
    {
        return f1+f2;
    }
    
    long double addLongDouble(long double f1 ,long double f2) const
    {
        return f1+f2;
    }
};
}
BCWASM_ABI(demo::Demo, init)
BCWASM_ABI(demo::Demo, addLong)
BCWASM_ABI(demo::Demo, addUlong)
BCWASM_ABI(demo::Demo, addFloat)
BCWASM_ABI(demo::Demo, addDouble)
BCWASM_ABI(demo::Demo, addLongDouble)
