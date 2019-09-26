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

char mapStrName[] = "mapstr";
typedef bcwasm::db::Map<mapStrName, string, string> MapStr;

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
            // map插入k-v数据
            void insertMap(const char* key, const char* value)
            {
                m_map.insert(key, value);
                /*    
                int nSize = m_map.size();
                bcwasm::println("m_map.size()：", nSize);
                if(4 == nSize)
                {
                    bcwasmThrow("over size:4");
                }*/
            }

            // 删除元素
            void delByKey(const char* key)
            {
                bcwasm::println("del m_map[", key, "]:", m_map[key]);
                m_map.del(key);
            }

            // 根据map的key获取value（用于非const）,如果key不存在，相当于insert
            int setValue(const char* key, const char* newValue)
            {
                string str = m_map[key];
                bcwasm::println("old m_map[", key, "]:", str);
                // map[]运算符返回引用
                m_map[key] = newValue;
                bcwasm::println("new m_map[", key, "]:", newValue);

				return 0;
            }

            // 获取map的长度
            int getMapSize() const
            {
                int nSize = m_map.getSize();
                bcwasm::println("m_map.size()：", nSize);
                return nSize;
            }

            // 根据map的key获取value(用于const)
            const char* getValueByKey(const char* key) const
            {
                string str = m_map.getConst(key);
                bcwasm::println("m_map[", key, "]:", str);

				char* buf = (char*)malloc(str.size() + 1);
				memset(buf, 0, str.size()+1);
				strcpy(buf, str.c_str());
				return buf;
            }

            // 遍历map(非const接口遍历)
            const char* traverseMap() 
            {
                string strRes = "{[";
                for (MapStr::Iterator iter = m_map.begin(); iter != m_map.end(); iter++)
                {
                    string key = iter->first();
                    string value = iter->second();
                    strRes += "\"" + key + "\":" + "\"" + value + "\",";
                }

                strRes += "]}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }

            // 遍历map（const接口遍历）
            const char* traverseMapConst() const
            {
                string strRes = "{";

                const std::set<string>& keys = m_map.getKeys();

                for(auto iter = keys.begin(); iter != keys.end(); ++iter)
                {
                    string strKey = *iter;
                    string strValue = m_map.getConst(strKey);
                    strRes += "\"" + strKey + "\":" + "\"" + strValue + "\",";
                }

                strRes += "}";

                char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
            }

        private:
            MapStr m_map;
    };
} // namespace demo

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(demo::FirstDemo, insertMap)
BCWASM_ABI(demo::FirstDemo, delByKey)
BCWASM_ABI(demo::FirstDemo, setValue)
BCWASM_ABI(demo::FirstDemo, getMapSize)
BCWASM_ABI(demo::FirstDemo, getValueByKey)
BCWASM_ABI(demo::FirstDemo, traverseMap)
BCWASM_ABI(demo::FirstDemo, traverseMapConst)

//bcwasm autogen begin
extern "C" { 
void insertMap(const char * key,const char * value) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.insertMap(key,value);
}
void delByKey(const char * key) {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.delByKey(key);
}
int setValue(const char * key,const char * newValue) {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.setValue(key,newValue);
}
int getMapSize() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getMapSize();
}
const char * getValueByKey(const char * key) {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.getValueByKey(key);
}
const char * traverseMap() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.traverseMap();
}
const char * traverseMapConst() {
demo::FirstDemo FirstDemo_bcwasm;
return FirstDemo_bcwasm.traverseMapConst();
}
void init() {
demo::FirstDemo FirstDemo_bcwasm;
FirstDemo_bcwasm.init();
}

}
//bcwasm autogen end