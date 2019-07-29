#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>


namespace bcwasm {
    class Hello : public Contract {
        public:
            Hello(){}
            virtual void init() {
                println("init success...");
            }
            BCWASM_EVENT(hello, const char *)
        public:
            void print(const char *msg)  {
                println(msg);
                BCWASM_EMIT_EVENT(hello, msg);
            }
    };

}

BCWASM_ABI(bcwasm::Hello, print)
////////////////bcwasm autogen begin
//////////////extern "C" { 
//////////////void print(const char * msg) {
//////////////bcwasm::Hello Hello_bcwasm;
//////////////Hello_bcwasm.print(msg);
//////////////}
//////////////void init() {
//////////////bcwasm::Hello Hello_bcwasm;
//////////////Hello_bcwasm.init();
//////////////}
//////////////
//////////////}
////////////////bcwasm autogen end
