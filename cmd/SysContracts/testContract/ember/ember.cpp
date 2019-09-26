//auto create contract
#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

#include "../../systemContract/util/util.hpp"

using namespace std;

namespace Ember
{
    class EmberContract : public bcwasm::Contract
    {
    public:
        EmberContract() {}

        /// 实现父类: bcwasm::Contract 的虚函数
        /// 该函数在合约首次发布时执行，仅调用一次
        void init()
        {
            util::registerContractFromInit("ember_contract","1.0.0.0");
            string addr = string("0x") + bcwasm::origin().toString();
            addBalance(addr.c_str(), 900000000000000000);
            bcwasm::println("EmberContract init success...");
        }

    public:
        void addBalance(const char* addr, unsigned long long balance)
        {
             unsigned long long curBalance = getBalance(addr) ;
            bcwasm::setState(addr, curBalance + balance);
        }

        void subBalance(const char* addr, unsigned long long balance)
        {
             unsigned long long curBalance = getBalance(addr) ;
             if (curBalance < balance)
             {
                 return;
             }

            bcwasm::setState(addr, curBalance -  balance);
        }

        unsigned long long getBalance(const char* addr) const
        {
            unsigned long long balance;
            bcwasm::getState(addr, balance);

            return balance;
        }
    };
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(Ember::EmberContract, getBalance)
BCWASM_ABI(Ember::EmberContract, addBalance)
BCWASM_ABI(Ember::EmberContract, subBalance)

//bcwasm autogen begin
extern "C" { 
void addBalance(const char * addr,unsigned long long balance) {
Ember::EmberContract EmberContract_bcwasm;
EmberContract_bcwasm.addBalance(addr,balance);
}
void subBalance(const char * addr,unsigned long long balance) {
Ember::EmberContract EmberContract_bcwasm;
EmberContract_bcwasm.subBalance(addr,balance);
}
unsigned long long getBalance(const char * addr) {
Ember::EmberContract EmberContract_bcwasm;
return EmberContract_bcwasm.getBalance(addr);
}
void init() {
Ember::EmberContract EmberContract_bcwasm;
EmberContract_bcwasm.init();
}

}
//bcwasm autogen end