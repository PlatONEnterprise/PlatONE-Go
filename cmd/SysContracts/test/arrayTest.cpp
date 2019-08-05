//auto create contract
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

char ArrStrName[] = "Arraystr";
typedef bcwasm::db::Array <ArrStrName, std::string, 20> ArrayStr;

#define RETURN_CHARARRAY(src, size) \
do \
{ \
    char *buf = (char *)malloc(size); \
    memset(buf, 0, size); \
    strcpy(buf, src); \
    return buf; \
} \
while(0)

namespace demo
{
    char onchain[] = "onchain";

    struct ResInfo {
		int code;
        string msg;
        string data;
	};

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
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)
        public:
            // array插入数据
            void setElem(int nPos, const char* Value)
            {
                int nSize = m_array.size();
                bcwasm::println("nSize:", nSize, ",nPos:", nPos);
                if(nPos >= nSize)
                {
                    bcwasm::println("已经超出范围");
                    return;
                }
                
                m_array[nPos] = Value;

                if(2 == nPos)
                {
                    bcwasmThrow("over size");
                }
            //    m_array.setConst(nPos, Value);
            }

            // 获取值
            const char* getValueByIndex(int index) const
            {
                if(index >= m_array.size() || index < 0)
                {
                    return "";
                }
                return m_array.getConst(index).c_str();
            }

            // 遍历array(在非const接口使用迭代器)
            const char* traverseArray()
            {
                string strRes = "{[";
                for (ArrayStr::Iterator iter = m_array.begin(); iter != m_array.end(); iter++)
                {
                    strRes += "\"" + *iter + "\",";
                }

                strRes += "]}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }

            // 遍历array（const接口遍历）
            const char* traverseArrayConst() const
            {
                string strRes = "{";

                for(int i = 0; i < m_array.size(); ++i)
                {
                    string strValue = m_array.getConst(i);
                    strRes += "\"" + strValue + "\",";
                }

                strRes += "}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }
        private:
            ArrayStr m_array;
    };
} // namespace demo

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::FirstDemo, setElem)
BCWASM_ABI(demo::FirstDemo, getValueByIndex)
BCWASM_ABI(demo::FirstDemo, traverseArray)
BCWASM_ABI(demo::FirstDemo, traverseArrayConst)

//bcwasm autogen begin
extern "C" { 
void setElem(int nPos,const char * Value) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.setElem(nPos,Value);
}
const char * getValueByIndex(int index) {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getValueByIndex(index);
}
const char * traverseArray() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.traverseArray();
}
const char * traverseArrayConst() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.traverseArrayConst();
}
void init() {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.init();
}

}
//bcwasm autogen end