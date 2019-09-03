//
// Created by zhou.yang on 2018/11/21.
//

#include "../unittest.hpp"
#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include "bcwasm/print.h"

#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>  
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

using namespace rapidjson;
using namespace bcwasm;
using namespace std;

namespace SystemContract
{
	struct RegisterInfo {
		std::string user_address;     // 用户账户地址
		std::string name;             // 用户名称
		std::string mobile;           // 手机号
		std::string email;            // 邮箱
		std::string remark;           // 个人说明
		unsigned int user_state;      // 平台用户状态：1审核中，2已激活，3已拒绝
		std::string auditor_address;  // 审核人的地址
	//	std::string roles;            // 角色请求列表
		std::vector<string> roles;   // 角色请求列表
	};

	struct ResultInfo {
		string code;
		string msg;
		string data;
	};

	enum USER_STATE {
		USER_STATE_AUDIT = 1,
		USER_STATE_ACTIVE = 2,
		USER_STATE_REJECT = 3
	};

    class UserRegister : public bcwasm::Contract
    {
        public:
            UserRegister(){}

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)

        public:
			//{"address":"0x33d253386582f38c66cb5819bfbdaad0910339b3","name":"xiaoluo","mobile":"13111111111","email":"luodahui@qq.com","roles":["a","b","c"],"remark":"平台用户申请"}
			int registerUser(const char* address, const char* name, const char* registJson) 
			{
                setState<string, string>(address, registJson);

                // 5.1 写入用户名关联的用户地址
				setState<string, string>(name, address);
                
				bcwasm::println("保存用户申请记录成功");
				return 0;
			}

			// 根据地址查询account信息
			const char* getAccountByAddress(const char* address) const
			{
				string strUserAddr = address;
				string userInfo = "";

				transform(strUserAddr.begin(), strUserAddr.end(), strUserAddr.begin(), ::tolower); 
                bcwasm::getState<string, string>(strUserAddr, userInfo);

				string code = "0";
				string msg = "succeed";
				if(userInfo.empty())
				{
					code = "1";
					msg = "The user does not exist in the UserRegister";
					userInfo = "\"\"";
				}
                return getResJson(code, msg, userInfo);
			}
			
			// 根据用户名查询account信息
			const char* getAccountByUsername(const char* UserName) const
			{
				string userInfo;
				string strUserAddr = "";
				// 根据用户名查询关联用户地址
				string strUserName = UserName;
				bcwasm::getState<std::string, std::string>(strUserName, strUserAddr);
				if("" == strUserAddr)
				{
				//	bcwasm::println("不存在用户：[", UserName, "],请检查");
					return getResJson("1", "The user does not exist in the UserRegister", "\"\"");
				}
				return getAccountByAddress(strUserAddr.c_str());
			}
			
		private:
			// 处理返回结果
			const char* getResJson(const string& code, const string& message, const string& dataInfo) const
			{
				string strRetJson = "{\"code\":" + code + ",";
				strRetJson += "\"msg\":\"" + message + "\",";
				strRetJson += "\"data\":" + dataInfo + "}";
				return strRetJson.c_str();
			}
    };
} // namespace SystemContract

TEST_CASE(test, UserRegister){
    {
        // 写入数据到db中
        string registJson = R"E({
        "user_address":"0xa0b21d5bcc6af4dda0579174941160b9eecb6918",
        "name":"user2",
        "mobile" : "1311111111111",
        "email" : "test1@qq.com",
        "remark" : "this is userRegister unittest",
        "user_state" : 1,
        "roles":["chainAdmin","nodeAdmin"]
        })E";

        SystemContract::UserRegister userReg;
        userReg.registerUser("0xa0b21d5bcc6af4dda0579174941160b9eecb6918", "user2", registJson.c_str());
        bcwasm::println("注册用户完成");
    }
    {
        // 从db中读取数据：根据地址查询用户申请数据
        SystemContract::UserRegister userReg;
        string strJson = userReg.getAccountByAddress("0xa0b21d5bcc6af4dda0579174941160b9eecb6918");
        bcwasm::println("get by user address:", strJson.c_str());
    }
    {
        // 从db中读取数据：根据用户名查询用户申请数据
        SystemContract::UserRegister userReg;
        string strJson = userReg.getAccountByUsername("user2");
        bcwasm::println("get by username:", strJson.c_str());
    }
}

UNITTEST_MAIN() {
    RUN_TEST(test, UserRegister)
}