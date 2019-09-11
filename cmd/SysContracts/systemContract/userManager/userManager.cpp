#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>  
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

#include "../common/util.hpp"

using namespace rapidjson;
using namespace bcwasm;
using namespace std;


const char* registAddr = "0x0000000000000000000000000000000000000011";

namespace SystemContract
{
    struct UserInfo
    {
        string user_address;	// 用户地址
        string name;//用户名称 
        string email; //用户邮箱 
        string mobile; //手机号
        unsigned int status;        // 是否禁用用户：[0:可用; 1:禁用; 2:删除] 
    };

    class UserManager : public bcwasm::Contract
    {
        public:
            UserManager(){}

            /// 实现父类: bcwasm::Contract 的虚函数
            /// 该函数在合约首次发布时执行，仅调用一次
            void init()
            {
                bcwasm::println("init success...");
                // 注册合约到合约管理合约
				DeployedContract reg(registAddr);
                reg.call("cnsRegisterFromInit", "__sys_UserManager", "1.0.0.0");

                string strOrigin = origin().toString();
                util::formatAddress(strOrigin);

                // 添加超级管理员
                UserInfo user;
                user.name = "root";
                user.user_address = strOrigin;
                user.status = 0;

                storeUserRecord(user);
                // 写入用户名对应的用户地址
                setState<string, string>(user.name, user.user_address);

                BCWASM_EMIT_EVENT(Notify, 0, "保存超级管理员成功!");
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)
            enum Code
            {
                SUCCESS,
                DESERIALIZE_ERROR,
                NAME_NULL,
                USER_EXIST,
                NO_PERMISSION,
                INVALID_USER,
                NO_APPLY,
                USER_STATUS_ERROR,
                NO_USER
            };

        public:
			//{"address":"0x33d253386582f38c66cb5819bfbdaad0910339b3","name":"xiaoluo","mobile":"13111111111","email":"luodahui@qq.com","status":0}
			int addUser(const char* userJson)
			{
                // 1. 反序列化用户信息
				UserInfo user;
				if (!deserializeDbUserRecord(userJson, user)) 
				{
				//	bcwasm::println("解析用户信息失败");
				    BCWASM_EMIT_EVENT(Notify, DESERIALIZE_ERROR, "解析用户信息失败");
                	return DESERIALIZE_ERROR;
				}	
                // 判断参数有效性
                if ("" == user.name) 
                {
                //    bcwasm::println("user name is invalid");
                    BCWASM_EMIT_EVENT(Notify, NAME_NULL, "user name is null");
                    return NAME_NULL;
                }

                // 2. 判断用户是否存在
                if(isStored(user.user_address))
                {
                //    bcwasm::println("user alrealdy exist");
                    BCWASM_EMIT_EVENT(Notify, USER_EXIST, "user alrealdy exist");
                    return USER_EXIST;
                }	
				
				string strOrigin = origin().toString();
                util::formatAddress(strOrigin);
                // 3. 判断调用者是否是管理员角色
                if(!isAdminRole(strOrigin))
                {
                //    bcwasm::println("无添加用户权限");
                    BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "无添加用户权限");
                    return NO_PERMISSION; 
                }
                // 4. 申请人是否是有效用户
				if(!isValidUser(strOrigin.c_str()))
				{
				//	bcwasm::println("调用者不是有效用户，申请失败");
                    BCWASM_EMIT_EVENT(Notify, INVALID_USER, "调用者不是有效用户，申请失败");
					return INVALID_USER;
				}

                // 5. 判断用户是否存在用户申请合约中，并且状态为审核通过状态
                int applyStatus = getUserApplyStatus(user.user_address);
                if(-1 == applyStatus)
                {
                //    bcwasm::println("用户申请记录不存在");
                    BCWASM_EMIT_EVENT(Notify, NO_APPLY, "用户申请记录不存在");
					return NO_APPLY;
                }

                // 1审核中，2已激活，3已拒绝
                if(2 != applyStatus)
                {
                //    bcwasm::println("用户状态为非审核通过状态，不能添加用户,当前状态为：[", applyStatus, "],请检查");
                    BCWASM_EMIT_EVENT(Notify, USER_STATUS_ERROR, "用户状态为非审核通过状态，不能添加用户");
					return USER_STATUS_ERROR;
                }

                // 6. 写入一条平台用户数据
                user.status = 0;  // 可用状态
				storeUserRecord(user);

                // 写入用户名对应的用户地址
                bcwasm::setState<std::string, std::string>(user.name, user.user_address);

				bcwasm::println("保存用户记录成功");
                BCWASM_EMIT_EVENT(Notify, SUCCESS, "保存用户记录成功");
                return SUCCESS;
            }

            // 将用户设置为可用状态
            int enable(const char* userAddr)
            {
                return setUserStatus(userAddr, 0);
            }
            // 将用户设置为禁用状态
            int disable(const char* userAddr)
            {
                if (isChainCreator(string(userAddr)))
                {
                   return NO_PERMISSION;
                }
                return setUserStatus(userAddr, 1);
            }

            // 将用户设置为删除用户
            int delUser(const char* userAddr)
            {
                return setUserStatus(userAddr, 2);
            }
            
            // 更新用户信息
            int update(const char* userAddr, const char* updateJson)
            {
                string strUserAddr = userAddr;
                util::formatAddress(strUserAddr);

                std::string db_value;
                bcwasm::getState(strUserAddr, db_value);
                if (db_value.empty()) 
                {
                //    bcwasm::println("user is not exist");
                    BCWASM_EMIT_EVENT(Notify,NO_USER , "user is not exist");
                    return NO_USER;
                }

                UserInfo userOld;
                 // 1. 反序列化用户信息
				if (!deserializeDbUserRecord(db_value, userOld)) 
				{
				//	bcwasm::println("解析用户信息失败");
                    BCWASM_EMIT_EVENT(Notify, DESERIALIZE_ERROR, "解析用户信息失败");
					return DESERIALIZE_ERROR;
				}	

                // 判断被更新用户是否是有效用户
                if(!isValidUser(userOld.user_address.c_str()))
                {
                //    bcwasm::println("no permission. invalid user:[", userOld.user_address, "]");
                    BCWASM_EMIT_EVENT(Notify, INVALID_USER, "invalid user");
                    return INVALID_USER;
                }

                UserInfo userUpdate;
                 // 反序列化用户信息
				if (!deserializeDbUserRecord(updateJson, userUpdate)) 
				{
				//	bcwasm::println("解析更新用户信息失败：[", updateJson, "]");
                    BCWASM_EMIT_EVENT(Notify, DESERIALIZE_ERROR, "解析更新用户信息失败");
					return DESERIALIZE_ERROR;
				}	

                // 2. 调用者是否是有效用户
                string strOrigin = origin().toString();
                util::formatAddress(strOrigin);
                if(!isValidUser(strOrigin.c_str()))
                {
                    BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "调用者不是有效用户");
                //     bcwasm::println("no permission. invalid user:[", strOrigin, "]");
                    return NO_PERMISSION;
                }

                bool permissionPass = false;
                // 3. 判断调用者权限
                if((0 == strOrigin.compare(userOld.user_address)) ||
                    (isAdminRole(strOrigin) && !isAdminRole(userOld.user_address)) || // 管理员可以编辑非管理员用户信息
                    isChainCreator(strOrigin)) // 链创建者可以编辑所有用户的信息 
                {
                    permissionPass = true;
                }
                
                if (!permissionPass) 
                {
                    BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "调用者没有权限");
                // /    bcwasm::println("no permission.");
                    return NO_PERMISSION;
                }

                // 账户地址和用户名不可更改，其他基本信息可编辑（email,mobile,status）
                // 更新email
				if(!userUpdate.email.empty())
				{
					userOld.email = userUpdate.email;
				}

                // 更新mobile
                if(!userUpdate.mobile.empty())
				{
					userOld.mobile = userUpdate.mobile;
				}

                // 更新status
                if(userUpdate.status != userOld.status)
				{
					userOld.status = userUpdate.status;
				}

                storeUserRecord(userOld);

			//	bcwasm::println("更新记录成功!");
                BCWASM_EMIT_EVENT(Notify, SUCCESS, "更新记录成功");
                return SUCCESS;
            }

            // 根据地址获取用户信息
            const char* getAccountByAddress(const char* address) const
            {
                string strUserAddr = address;
                util::formatAddress(strUserAddr);

                std::string db_value;
                bcwasm::getState(strUserAddr, db_value);

                string code = "0";
				string msg = "succeed";
				if(db_value.empty())
				{
					code = "1";
					msg = "The user does not exist in the useManager contract";
                    db_value = "\"\"";
				}

                return getResJson(code, msg, db_value);
            }
            // 根据用户名获取用户信息
            const char* getAccountByName(const char* name) const
            {
                string db_value = "";
                string strAddress = "";
                string strName = name;

                bcwasm::getState(strName, strAddress);
                if("" == strAddress)
				{
					return getResJson("1", "The user does not exist in the useManager contract", "\"\"");
				}

                return getAccountByAddress(strAddress.c_str());
            }
			
            // 是否是有效用户
            int isValidUser(const char* userAddr) const
            {
                string strUserAddr = userAddr;
                util::formatAddress(strUserAddr);

                std::string db_value;
                bcwasm::getState(strUserAddr, db_value);
                
                if (db_value.empty()) 
                {
                    bcwasm::println("user is not exist in the userManager contract");
                    return 0;
                }

                Document document; 
                document.Parse<0>(db_value.c_str()); 

                if(!document.HasMember("status"))
                {
                //    bcwasm::println("user is not ValidUser");
                    return 0;
                }
               
                Value& status = document["status"];  
                int iStatus = status.GetInt();
            //    bcwasm::println("status:", iStatus);
                // 用户处于非可用状态
                if(0 != iStatus)
                {
                //    bcwasm::println("user is not ValidUser");
                    return 0;
                }
                
                return 1;
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

            // 设置用户状态
            int setUserStatus(const char* userAddr, unsigned int status)
            {
                string strUserAddr = userAddr;
				util::formatAddress(strUserAddr);

                // 1. 判断用户是否存在
                std::string db_value;
                bcwasm::getState(strUserAddr, db_value);
                if (db_value.empty()) 
                {
                    BCWASM_EMIT_EVENT(Notify, NO_USER, "user is not exist");
                //    bcwasm::println("user is not exist");
                    return NO_USER;
                }

                string strOrigin = origin().toString();
                util::formatAddress(strOrigin);
                // 3. 判断调用者是否是管理员角色
                if(!isAdminRole(strOrigin))
                {
                    BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "调用者为非管理员，没有更新用户状态权限");
                //    bcwasm::println("调用者为非管理员，没有更新用户状态权限");
                    return NO_PERMISSION; 
                }

                // 4. 调用者是否是有效用户
				if(!isValidUser(strOrigin.c_str()))
				{
                    BCWASM_EMIT_EVENT(Notify, INVALID_USER, "调用者不是有效用户，更新用户状态失败");
				//	bcwasm::println("调用者不是有效用户，更新用户状态失败");
					return INVALID_USER;
				}

                Document document; 
				document.Parse<0>(db_value.c_str()); 	
				
				//2.取值
				if(document.HasMember("status"))
				{
					Value& statusValue = document["status"];  
                    int iStatus = statusValue.GetInt();
                    if(2 == iStatus)
                    {
                        // 已删除用户不能变更状态
                        BCWASM_EMIT_EVENT(Notify, USER_STATUS_ERROR, "已删除用户不能变更状态");
                        return USER_STATUS_ERROR;
                    }
                //    bcwasm::println("status:", iStatus);
                    if(status != iStatus)
                    {
                        statusValue.SetInt(status);
                        StringBuffer buffer;
                        Writer<StringBuffer> writer(buffer);
                        document.Accept(writer);
                        std::string db_user_record = buffer.GetString();

                        bcwasm::setState<std::string, std::string>(strUserAddr, db_user_record);
                    }
				}
                BCWASM_EMIT_EVENT(Notify, SUCCESS, "用户状态更新成功");
                return SUCCESS;
            }

            // 判断申请人/审核人是否是管理员角色
			bool isAdminRole(const string& address) 
			{
				DeployedContract a(registAddr);
                string strRoleMageAddr = a.callString("getContractAddress", "__sys_RoleManager", "latest");

				DeployedContract role_c(strRoleMageAddr);

                string strResult = role_c.callString("getRolesByAddress", address.c_str());
				if(-1 == strResult.find("Admin") && -1 == strResult.find("chainCreator"))
				{
				//	bcwasm::println("调用者权限不合法，合约执行结果失败");
					return false;
				} 
				return true;
			}

            // 判断用户是否是链创建者
            bool isChainCreator(const string& address) 
            {
                DeployedContract a(registAddr);
                string strRoleMageAddr = a.callString("getContractAddress", "__sys_RoleManager", "latest");
				DeployedContract role_c(strRoleMageAddr);

                string strResult = role_c.callString("getRolesByAddress", address.c_str());
				if(-1 == strResult.find("chainCreator"))
				{
					return false;
				} 
				return true;
            }

            // 获取用户申请状态
            int getUserApplyStatus(const string& address)
            {
                DeployedContract a(registAddr);
                string strRegMageAddr = a.callString("getContractAddress", "__sys_UserRegister", "latest");

				DeployedContract reg(strRegMageAddr);
				int ret = reg.callInt64("getStatusByAddress", address.c_str());
				
				return ret;
            }
            // 获取已经保存过的记录
			void getUserRecord(const std::string& address, std::string& db_user_record) 
			{
				 bcwasm::getState<std::string, std::string>(address, db_user_record);
			}
        	// 判断用户是否保存
			bool isStored(std::string& address) 
			{
				std::string db_user_record;
				getUserRecord(address, db_user_record);

				if (db_user_record.empty()) 
				{
				//	bcwasm::println("user:[", address, "] not found!");
					return false;
				}
				return true;
			}
            // 保存用户记录
            int storeUserRecord(UserInfo & user)
            {
                std::string db_user_record;
				serializeDbUserRecord(user, db_user_record);
				bcwasm::setState<std::string, std::string>(user.user_address, db_user_record);

                return 0;
            }
        	// 序列化用户信息
			void serializeDbUserRecord(UserInfo& user, string& db_user_record) 
			{
                // 当value的参数是字符串变量，且比较特殊时，如地址：0x.....,会导致rapidjson吐出空字符串
                if("" == user.mobile)
                {
                    user.mobile = "0";
                }
                db_user_record = "{";
                db_user_record += "\"address\":\"" + user.user_address + "\",";
                db_user_record += "\"name\":\"" + user.name + "\",";
                db_user_record += "\"mobile\":\"" + user.mobile + "\",";
                db_user_record += "\"email\":\"" + user.email + "\",";
                db_user_record += "\"status\":" + to_string(user.status) + "}";
                
                /*
				Document doc;
				doc.SetObject();
				Document::AllocatorType &allocator = doc.GetAllocator();

				Value value(kStringType);
				value.SetString(user.user_address.c_str(), user.user_address.size(), allocator);
				doc.AddMember("address", value, allocator);

				value.SetString(user.name.c_str(), user.name.size(), allocator);
				doc.AddMember("name", value, allocator);

				value.SetString(user.mobile.c_str(), user.mobile.size(), allocator);
				doc.AddMember("mobile", value, allocator);

				value.SetString(user.email.c_str(), user.email.size(), allocator);
				doc.AddMember("email", value, allocator);

				value.SetInt(user.status);
				doc.AddMember("status", value, allocator);

				StringBuffer buffer;
				Writer<StringBuffer> writer(buffer);
				doc.Accept(writer);
				
				db_user_record = buffer.GetString();*/

				bcwasm::println("用户数据===:", db_user_record.c_str());
			}

			// 反序列化用户信息
			bool deserializeDbUserRecord(const std::string& db_user_record, UserInfo& user) const 
			{
				Document doc;
				doc.Parse<kParseDefaultFlags>(db_user_record.c_str());
				Document::AllocatorType &allocator = doc.GetAllocator();

				if (doc.HasParseError()) 
				{
					ParseErrorCode code = doc.GetParseError();
					return false;
				}

				Value::MemberIterator itr = doc.FindMember("address");
				if (itr == doc.MemberEnd()) 
				{
				//	bcwasm::println("没找到对应的 address");
				}
				else 
				{
					user.user_address = itr->value.GetString();
                    // 转换成小写
					util::formatAddress(user.user_address);
				}

				itr = doc.FindMember("name");
				if (itr == doc.MemberEnd()) 
				{
				//	bcwasm::println("没找到对应的 name");
				}
				else 
				{
					user.name = itr->value.GetString();
				}

				itr = doc.FindMember("mobile");
				if (itr == doc.MemberEnd()) 
				{
				//	bcwasm::println("没找到对应的 mobile");
				}
				else 
				{
					user.mobile = itr->value.GetString();
				}

				itr = doc.FindMember("email");
				if (itr == doc.MemberEnd()) 
				{
				//	bcwasm::println("没找到对应的 email");
				}
				else 
				{
					user.email = itr->value.GetString();
				}

				itr = doc.FindMember("status");
				if (itr == doc.MemberEnd()) 
				{
				//	bcwasm::println("没找到对应的 status");
				}
				else 
				{
					user.status = itr->value.GetInt();
				}
				
				return true;
			}
    };
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::UserManager, addUser)
BCWASM_ABI(SystemContract::UserManager, enable)
BCWASM_ABI(SystemContract::UserManager, disable)
BCWASM_ABI(SystemContract::UserManager, delUser)
BCWASM_ABI(SystemContract::UserManager, getAccountByAddress)
BCWASM_ABI(SystemContract::UserManager, getAccountByName)
BCWASM_ABI(SystemContract::UserManager, isValidUser)
BCWASM_ABI(SystemContract::UserManager, update)
BCWASM_ABI(SystemContract::UserManager, test_isValidUser)
BCWASM_ABI(SystemContract::UserManager, test_hasRole)
BCWASM_ABI(SystemContract::UserManager, update)

//bcwasm autogen begin
extern "C" { 
int addUser(const char * userJson) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.addUser(userJson);
}
int enable(const char * userAddr) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.enable(userAddr);
}
int disable(const char * userAddr) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.disable(userAddr);
}
int delUser(const char * userAddr) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.delUser(userAddr);
}
int update(const char * userAddr,const char * updateJson) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.update(userAddr,updateJson);
}
const char * getAccountByAddress(const char * address) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.getAccountByAddress(address);
}
const char * getAccountByName(const char * name) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.getAccountByName(name);
}
int isValidUser(const char * userAddr) {
SystemContract::UserManager UserManager_bcwasm;
return UserManager_bcwasm.isValidUser(userAddr);
}
void init() {
SystemContract::UserManager UserManager_bcwasm;
UserManager_bcwasm.init();
}

}
//bcwasm autogen end