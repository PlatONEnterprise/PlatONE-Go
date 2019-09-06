#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

namespace demo {
    class FirstDemo : public bcwasm::Contract
    {
        public:
            FirstDemo(){}

            /// 实现父类: bcwasm::Contract 的虚函数
            /// 该函数在合约首次发布时执行，仅调用一次
            void init() 
            {
                bcwasm::println("init success...");
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)

        public:
            void invokeNotify(const char *msg)
            {    
                // 定义状态变量
                bcwasm::setState("NAME_KEY", std::string(msg));
                // 日志输出
                bcwasm::println("into invokeNotify...");
                // 事件返回
                BCWASM_EMIT_EVENT(Notify, 0, "Insufficient value for the method.");
            }

            const char* getName() const 
            {
                std::string value;
                bcwasm::getState("NAME_KEY", value);
                // 读取合约数据并返回
                return value.c_str();
            }
    };
}

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::FirstDemo, invokeNotify)
BCWASM_ABI(demo::FirstDemo, getName)
//bcwasm autogen begin
extern "C" { 
void invokeNotify(const char * msg) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.invokeNotify(msg);
}
const char * getName() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getName();
}
void init() {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.init();
}

}
//bcwasm autogen end