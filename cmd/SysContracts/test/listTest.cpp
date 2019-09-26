#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>  
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

using namespace rapidjson;
using namespace bcwasm;
using namespace std;
// 审核中
char listStrName[] = "listStr";

typedef bcwasm::db::List<listStrName, string> ListStr;


namespace demo
{
    char onchain[] = "onchain";

    class FirstDemo : public bcwasm::Contract
    {
        public:
            FirstDemo() 
			{
			}

            /// 实现父类: bcwasm::Contract 的虚函数
            /// 该函数在合约首次发布时执行，仅调用一次
            void init()
            {
                bcwasm::println("init success...");
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)

        public:
            // 添加
            void insert(const char* value)
            {
                m_list.push(value);
                /*    
                int nSize = m_list.size();
                bcwasm::println("m_list.size()：", nSize);
                if(4 == nSize)
                {
                    bcwasmThrow("over size:4");
                }*/
            }

            // 删除
            void del(int nIndex)
            {
                m_list.del(nIndex);
            }

            // 修改
            void setValue(int nIndex, const char* newValue)
            {
                if(nIndex < 0 || nIndex >= m_list.size())
                {
                    bcwasmThrow("over size");
                    return;
                }
                m_list[nIndex] = newValue;
            }

            // 获取值
            const char* getValue(int index) const
            {
                string str = m_list.getConst(index);
                bcwasm::println("m_list[", index, "]:", str);
                return str.c_str();
            }

            // 获取长度
            int getSize() const
            {
                int nSize = m_list.size();
                bcwasm::println("m_list.size()：", nSize);
                return nSize;
            }

             // 遍历List(在非const接口可使用迭代器)
            const char* traverseList()
            {
                string strRes = "{[";
                for (ListStr::Iterator iter = m_list.begin(); iter != m_list.end(); iter++)
                {
                    strRes += "\"" + *iter + "\",";
                }

                strRes += "]}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }

            // 遍历（const接口使用）
            const char* traverseListConst() const
            {
                string strRes = "{";
                for(int i = 0; i < m_list.size(); ++i)
                {
                    string str = m_list.getConst(i);
                    strRes += "\"" + str + "\",";
                }
                strRes += "}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }

        private:
            ListStr m_list;
    };
} // namespace demo

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::FirstDemo, insert)
BCWASM_ABI(demo::FirstDemo, del)
BCWASM_ABI(demo::FirstDemo, setValue)
BCWASM_ABI(demo::FirstDemo, getValue)
BCWASM_ABI(demo::FirstDemo, getSize)
BCWASM_ABI(demo::FirstDemo, traverseListConst)
//bcwasm autogen begin
extern "C" { 
void insert(const char * value) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.insert(value);
}
void del(int nIndex) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.del(nIndex);
}
void setValue(int nIndex,const char * newValue) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.setValue(nIndex,newValue);
}
const char * getValue(int index) {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getValue(index);
}
int getSize() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getSize();
}
const char * traverseListConst() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.traverseListConst();
}
void init() {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.init();
}

}
//bcwasm autogen end