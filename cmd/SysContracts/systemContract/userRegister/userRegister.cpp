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

// 审核中
char listAudit[] = "listAudit";
// 已激活(审核通过)
char listActive[] = "listActive";
// 拒绝
char listReject[] = "listReject";

typedef bcwasm::db::List<listAudit, string> ListAudit;
typedef bcwasm::db::List<listActive, string> ListActive;
typedef bcwasm::db::List<listReject, string> ListReject;

const char* registAddr = "0x0000000000000000000000000000000000000011";

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

	enum UserState {
		USER_STATE_AUDIT = 1,
		USER_STATE_ACTIVE = 2,
		USER_STATE_REJECT = 3
	};
	enum RegisterUserState_Code
	{
		R_SUCCESS = 0,
		DESERIALIZE_REGISTER_ERROR,
		NO_PERMISSION,
		RECORD_EXIT,
		NAME_NULL,
		NAME_APPLIED,
		DESERIALIZE_DATA_ERROR,
		NAME_EXIT,
		STORE_ERROR
	};
	enum Approve_Code
	{
		A_SUCCESS = 0,
		STATE_ERROR,
		NO_INFO,
		INVALID_USER,
		A_NO_PERMISSION,
		DESERIALIZE_INFO_ERROR,
		STARE_RECORD_ERROR,
		ADD_ROLE_ERROR,
		REGISTER_STR_LEN_ERROR
	};

    class UserRegister : public bcwasm::Contract
    {
        public:
            UserRegister(){}

            /// 实现父类: bcwasm::Contract 的虚函数
            /// 该函数在合约首次发布时执行，仅调用一次
            void init()
            {
                bcwasm::println("init success...");
				// 注册合约到合约管理合约
				DeployedContract reg(registAddr);
                reg.call("cnsRegisterFromInit", "__sys_UserRegister", "1.0.0.0");
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)

        public:
			//{"address":"0x33d253386582f38c66cb5819bfbdaad0910339b3","name":"xiaoluo","mobile":"13111111111","email":"luodahui@qq.com","roles":["a","b","c"],"remark":"平台用户申请"}
			int registerUser(const char* registJson)
			{
                std::string regStr(registJson);
                if (800 < regStr.size())
                {
                    BCWASM_EMIT_EVENT(Notify, REGISTER_STR_LEN_ERROR, "用户信息参数json字符串长度>800, 注册失败");
                    return REGISTER_STR_LEN_ERROR;
                }

				// 1. 反序列化用户申请信息(判断用户信息)
				RegisterInfo reg_user;
				if (!deserializeDbRegisterRecord(regStr, reg_user))
				{
				//	bcwasm::println("解析用户注册信息失败，获取状态失败");
					BCWASM_EMIT_EVENT(Notify, DESERIALIZE_REGISTER_ERROR, "解析用户注册信息失败，获取状态失败");
					return DESERIALIZE_REGISTER_ERROR;
				}
			
				string strOrigin = origin().toString();
				util::formatAddress(strOrigin);
				// 自己发起申请
				if(reg_user.user_address != strOrigin)
				{
				//	bcwasm::println("只有自己可申请自己为平台用户");
					BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "只有自己可申请自己为平台用户");
					return NO_PERMISSION;
				}

				// 判断用户申请记录是否存在(一个用户只能申请一次)
				if(isRegistered(reg_user.user_address.c_str()))
				{
				//	bcwasm::println("申请记录已存在");
					BCWASM_EMIT_EVENT(Notify, RECORD_EXIT, "申请记录已存在");
					return RECORD_EXIT;
				}
				
				// 判断用户信息是否为空
				if("" == reg_user.name)
				{
				//	bcwasm::println("用户名不能为空");
					BCWASM_EMIT_EVENT(Notify, NAME_NULL, "用户名不能为空");
					return NAME_NULL;
				}


				// 4. 判断平台用户的用户名是否已经申请
				// 4.1 获取
				string strUserAddr = "";
				bcwasm::getState<std::string, std::string>(reg_user.name, strUserAddr);
				if("" != strUserAddr)
				{
				//	bcwasm::println("平台用户的用户名已申请，申请失败");
					BCWASM_EMIT_EVENT(Notify, NAME_APPLIED, "平台用户的用户名已申请或已审核通过，申请失败");
					return NAME_APPLIED;
				}

				// 4.2 获取平台用户管理合约UserManage地址
				DeployedContract a(registAddr);
                string strUserMageAddr = a.callString("getContractAddress", "__sys_UserManager", "latest");
				// 调用UserManage合约getAccountByName接口判断用户名是否存在
				DeployedContract b(strUserMageAddr);
				string strResult = b.callString("getAccountByName", reg_user.name.c_str());

				ResultInfo res;
				if(!desResultInfo(strResult, res))
				{
				//	bcwasm::println("解析用户数据失败");
					BCWASM_EMIT_EVENT(Notify, DESERIALIZE_DATA_ERROR, "解析用户数据失败");
					return DESERIALIZE_DATA_ERROR;
				}

				if("0" == res.code)
				{
				//	bcwasm::println("平台用户的用户名已经存在");
					BCWASM_EMIT_EVENT(Notify, NAME_EXIT, "平台用户的用户名已经存在");
					return NAME_EXIT;
				}

				// 5.写入一条平台用户申请记录（状态为审核中）
				reg_user.user_state = USER_STATE_AUDIT;
				int ret = storeRegisterRecord(reg_user, reg_user.user_address);
				if (ret != 0) 
				{
				//	bcwasm::println("写入用户申请记录失败");
					BCWASM_EMIT_EVENT(Notify, STORE_ERROR, "写入用户申请记录失败");
					return STORE_ERROR;
				}

				// 5.1 写入用户名关联的用户地址
				bcwasm::setState<std::string, std::string>(reg_user.name, reg_user.user_address);

				// 6.保存待审核用户地址（用于查询相应用户状态的所有用户信息）
				m_listAudit.push(reg_user.user_address);

			//	bcwasm::println("保存用户申请记录成功");
				BCWASM_EMIT_EVENT(Notify, R_SUCCESS, "保存用户申请记录成功");
				return R_SUCCESS;
			}

			// 审核接口
			int approve(const char* userAddress, int auditStatus)
			{
				// 1.检查入参
				if (auditStatus != USER_STATE_REJECT && auditStatus != USER_STATE_ACTIVE) 
				{
				//	bcwasm::println("用户状态不合法");
					BCWASM_EMIT_EVENT(Notify, STATE_ERROR, "用户状态不合法");
					return STATE_ERROR;
				}

				// 2.校验被审核用户是否已经被申请
				string db_reg_record;
				string strUserAddr = userAddress;
				util::formatAddress(strUserAddr);

				getRegisterRecord(strUserAddr, db_reg_record);
				if (db_reg_record.empty()) 
				{
				//	bcwasm::println("审核的用户地址:[", strUserAddr, "] 不存在!");
					BCWASM_EMIT_EVENT(Notify, NO_INFO, "用户信息不存在");
					return NO_INFO;
				}

				// 3. 检查被审核用户的状态，非审核中不可被审核
				int tmp_state = getStatusByAddress(strUserAddr.c_str());
				if (-1 == tmp_state) 
				{
				//	bcwasm::println("获取被审核者用户状态失败");
					BCWASM_EMIT_EVENT(Notify, STATE_ERROR, "获取被审核者用户状态失败");
					return STATE_ERROR;
				}

				if (tmp_state != USER_STATE_AUDIT) 
				{
				//	bcwasm::println("被审核的用户为非审核中");
					BCWASM_EMIT_EVENT(Notify, STATE_ERROR, "被审核的用户为非审核中");
					return STATE_ERROR;
				}

				string strOrigin = origin().toString(); 
				util::formatAddress(strOrigin);
				// 4 判断审核人是否是有效用户
				if (!isValidUser(strOrigin))
				{
				//	bcwasm::println("审核人不是有效用户，审核失败");
					BCWASM_EMIT_EVENT(Notify, INVALID_USER, "审核人不是有效用户，审核失败");
					return INVALID_USER;    
				}	

				// 5.判断审核人是否是管理员角色
				if(!isAdminRole(strOrigin))
				{
				//	bcwasm::println("审核人没有管理员权限，审核失败");
					BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, "审核人没有管理员权限，审核失败");
					return NO_PERMISSION;
				}
							
				// 6.从数据库获取用户申请记录，修改用户状态
				RegisterInfo reg_user;
				if (!deserializeDbRegisterRecord(db_reg_record, reg_user)) 
				{
				//	bcwasm::println("解析用户注册信息失败");
					BCWASM_EMIT_EVENT(Notify, DESERIALIZE_INFO_ERROR, "解析用户注册信息失败");
					return DESERIALIZE_INFO_ERROR;
				}
				reg_user.user_state = auditStatus;
				reg_user.auditor_address = strOrigin;
				bcwasm::println("reg_user.user_state:", reg_user.user_state);

				// 修改审核信息
				if (0 != storeRegisterRecord(reg_user, reg_user.user_address)) 
				{
				//	bcwasm::println("修改用户审核记录失败");
					BCWASM_EMIT_EVENT(Notify, STARE_RECORD_ERROR, "修改用户审核记录失败");
					// rollback
					bcwasmThrow("修改用户审核记录失败");
					return STARE_RECORD_ERROR;
				}
				
				// 7. 新增平台用户--只有审核通过才新增
				if (USER_STATE_ACTIVE == auditStatus) 
				{		
					// 获取平台用户管理合约UserManage地址
					DeployedContract a(registAddr);
					string strUserMageAddr = a.callString("getContractAddress", "__sys_UserManager", "latest");
					
					// 调用UserManage合约addUser接口插入用户信息
					DeployedContract b(strUserMageAddr.c_str());
					int nRet = b.callInt64("addUser", db_reg_record.c_str());
					if(0 != nRet)
					{
					//	bcwasm::println("新增用户到平台用户管理合约失败");
						BCWASM_EMIT_EVENT(Notify, STARE_RECORD_ERROR, "新增用户到平台用户管理合约失败");
						// rollback
						bcwasmThrow("新增用户到平台用户管理合约失败");
						return STARE_RECORD_ERROR;
					}

					// 检查是否附带申请角色
					int nRoles = reg_user.roles.size();
					if( nRoles > 0)
					{
						string strRoles = "[";
						for(int i = 0; i < nRoles; ++i)
						{
							strRoles += "\"" + reg_user.roles[i] + "\"";
							if( i < nRoles -1 )
							{
								strRoles += ",";
							}
						}
						strRoles += "]";

						int nRet = setUserRole(reg_user.name, reg_user.user_address, strRoles);
						if(0 != nRet)
						{
						//	bcwasm::println("添加用户角色失败");
							BCWASM_EMIT_EVENT(Notify, ADD_ROLE_ERROR, "添加用户角色失败");
							// rollback
							bcwasmThrow("添加用户角色失败");
							return ADD_ROLE_ERROR;
						}
					}
					
					// 保存审核通过的用户账户，并从待审核中删除
					m_listActive.push(strUserAddr);
				}
				else
				{
					// 审核不通过
				//	bcwasm::println("用户申请被拒绝");
					delState(reg_user.name);
					m_listReject.push(strUserAddr);
				}

				// 8. 从审核中列表中删除用户地址
				m_listAudit.del(strUserAddr);
				BCWASM_EMIT_EVENT(Notify, A_SUCCESS, "审核成功");
				return A_SUCCESS;
			}
			
			// 根据地址查询account信息
			const char* getAccountByAddress(const char* address) const
			{
				string strUserAddr = address;
				string userInfo = "";
 
				util::formatAddress(strUserAddr);

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
			
			// 分页查询返回某个status的所有用户
			const char* getAccountsByStatus(int pageNum, int pageSize, int accountStatus) const
			{
				size_t size = 0;
				string userInfos = "";
				string code = "0";
				string message = "succeed";
				if(USER_STATE_AUDIT == accountStatus)
				{
					size = m_listAudit.size();
					if(0 < size)
					{
						unsigned int startIndex = pageNum * pageSize;
            			unsigned int endIndex = startIndex + pageSize - 1;
						if (startIndex >= size) 
						{
							code = "1";
							message = "Adjust pageNum and pageSize";
						//	bcwasm::println("size:", size, ",请调整pageNum和pageSize");
						//	return "";
						}

						if (endIndex >= size) 
						{
							endIndex = size - 1;
						}

						for (int i = startIndex; i <= endIndex; i++) 
						{
							string strUserAddr = m_listAudit.getConst(i);
							string strUserInfo = "";
                			getState<string, string>(strUserAddr, strUserInfo);

							userInfos += strUserInfo;
							if(i != endIndex)
							{
								userInfos += ",";
							}
						}
					}
				}
				else if(USER_STATE_ACTIVE == accountStatus)
				{
					size = m_listActive.size();
					if(0 < size)
					{
						unsigned int startIndex = pageNum * pageSize;
            			unsigned int endIndex = startIndex + pageSize - 1;
						if (startIndex >= size) 
						{
						//	bcwasm::println("size:", size, ",请调整pageNum和pageSize");
						//	return "";
							code = "1";
							message = "Adjust pageNum and pageSize";
						}

						if (endIndex >= size) 
						{
							endIndex = size - 1;
						}

						for (int i = startIndex; i <= endIndex; i++) 
						{
							string strUserAddr = m_listActive.getConst(i);
							string strUserInfo = "";
							getState<string, string>(strUserAddr, strUserInfo);
							userInfos += strUserInfo;
							if(i != endIndex)
							{
								userInfos += ",";
							}
						}
					}
				}
				else if(USER_STATE_REJECT == accountStatus)
				{
					size = m_listReject.size();
					if(0 < size)
					{
						unsigned int startIndex = pageNum * pageSize;
            			unsigned int endIndex = startIndex + pageSize - 1;
						if (startIndex >= size) 
						{
						//	bcwasm::println("size:", size, ",请调整pageNum和pageSize");
						//	return "";
							code = "1";
							message = "Adjust pageNum and pageSize";
						}

						if (endIndex >= size) 
						{
							endIndex = size - 1;
						}

						for (int i = startIndex; i <= endIndex; i++) 
						{
							string strUserAddr = m_listReject.getConst(i);
							string strUserInfo = "";
							getState<string, string>(strUserAddr, strUserInfo);
							userInfos += strUserInfo;
							// 拼装json
							if(i != endIndex)
							{
								userInfos += ",";
							}
						}
					}
				}
				else
				{
				//	bcwasm::println("状态用户不存在：[", accountStatus, "],请检查");
					code = "2";
					message = "The user status for the query does not exist in the UserRegister";
					userInfos = "\"\"";
				}

				if("0" == code)
				{
					if(0 == size)
					{
						code = "3";
						message = "user information is empty in the UserRegister";
						userInfos = "\"\"";
					}
					else if(size > 1)
					{
						userInfos = "[" + userInfos + "]";
					}
				}

				string strRes = getResJson(code, message, userInfos);
				char* buf = (char*)malloc(strRes.size() + 1);
				memset(buf, 0, strRes.size()+1);
				strcpy(buf, strRes.c_str());
				return buf;
			//	return getResJson(code, message, userInfos);
			}

			// 获取平台用户申请状态(审核中/审核通过/拒绝)
			int getStatusByAddress(const char* address) const 
			{
				string strUserAddr = address;
				util::formatAddress(strUserAddr);
				
				std::string db_reg_record;
				getRegisterRecord(strUserAddr, db_reg_record);

				RegisterInfo reg_user;
				bool ret = deserializeDbRegisterRecord(db_reg_record, reg_user);
				if (!ret) 
				{
					bcwasm::println("解析用户注册信息失败，获取状态失败");
					return -1;
				}
				
				return reg_user.user_state;
			}
		private:
			// 序列化用户信息
			void serializeDbRegisterRecord(RegisterInfo& reg_user, std::string& db_reg_record) 
			{
				db_reg_record = "{";
				util::formatAddress(reg_user.user_address);
                db_reg_record += "\"address\":\"" + reg_user.user_address + "\",";
                db_reg_record += "\"name\":\"" + reg_user.name + "\",";
                db_reg_record += "\"mobile\":\"" + reg_user.mobile + "\",";
                db_reg_record += "\"email\":\"" + reg_user.email + "\",";
				db_reg_record += "\"remark\":\"" + reg_user.remark + "\",";
				db_reg_record += "\"user_state\":" + to_string(reg_user.user_state) + ",";
				util::formatAddress(reg_user.auditor_address);
				db_reg_record += "\"auditor_address\":\"" + reg_user.auditor_address + "\",";

				string strRoles = "[";
				int nRoles = reg_user.roles.size();
				for(int i = 0; i < nRoles; ++i)
				{
					strRoles += "\"" + reg_user.roles[i] + "\"";
					if( i < nRoles -1 )
					{
						strRoles += ",";
					}
				}
				strRoles += "]";

                db_reg_record += "\"roles\":" + strRoles + "}";
			/*	
				Document doc;
				doc.SetObject();
				Document::AllocatorType &allocator = doc.GetAllocator();

				Value value(kStringType);
				value.SetString(reg_user.user_address.c_str(), reg_user.user_address.size(), allocator);
				doc.AddMember("address", value, allocator);

				value.SetString(reg_user.name.c_str(), reg_user.name.size(), allocator);
				doc.AddMember("name", value, allocator);

				value.SetString(reg_user.mobile.c_str(), reg_user.mobile.size(), allocator);
				doc.AddMember("mobile", value, allocator);

				value.SetString(reg_user.email.c_str(), reg_user.email.size(), allocator);
				doc.AddMember("email", value, allocator);

				value.SetString(reg_user.remark.c_str(), reg_user.remark.size(), allocator);
				doc.AddMember("remark", value, allocator);

				value.SetInt(USER_STATE_AUDIT);
				doc.AddMember("user_state", value, allocator);

				value.SetString(reg_user.auditor_address.c_str(), reg_user.auditor_address.size(), allocator);
				doc.AddMember("auditor_address", value, allocator);

				Value rolesArr(kArrayType);
				for(const auto& key : reg_user.roles)
				{
					Value role(key.c_str(), allocator);
					rolesArr.PushBack(role, allocator);
				}
				doc.AddMember("roles", rolesArr, allocator);

				StringBuffer buffer;
				Writer<StringBuffer> writer(buffer);
				doc.Accept(writer);
				
				db_reg_record = buffer.GetString();*/

				bcwasm::println("用户申请记录：", db_reg_record.c_str());
			}

			// 反序列化用户信息
			bool deserializeDbRegisterRecord(const std::string& db_reg_record, RegisterInfo& reg_user) const 
			{
				Document doc;
				doc.Parse<kParseDefaultFlags>(db_reg_record.c_str());
				Document::AllocatorType &allocator = doc.GetAllocator();

				if (doc.HasParseError()) 
				{
					ParseErrorCode code = doc.GetParseError();
					bcwasm::println("解析用户注册信息失败，错误码：", code);
					return false;
				}

				Value::MemberIterator itr = doc.FindMember("address");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 address");
				}
				else 
				{
					reg_user.user_address = itr->value.GetString();
					// 转换成小写
					util::formatAddress(reg_user.user_address);
				}

				itr = doc.FindMember("name");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 name");
				}
				else 
				{
					reg_user.name = itr->value.GetString();
				}

				itr = doc.FindMember("mobile");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 mobile");
				}
				else 
				{
					reg_user.mobile = itr->value.GetString();
				}

				itr = doc.FindMember("email");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 email");
				}
				else 
				{
					reg_user.email = itr->value.GetString();
				}

				itr = doc.FindMember("remark");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 remark");
				}
				else 
				{
					reg_user.remark = itr->value.GetString();
				}

				itr = doc.FindMember("user_state");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 user_state");
				}
				else 
				{
					reg_user.user_state = itr->value.GetInt();
				}
				
				itr = doc.FindMember("auditor_address");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 auditor_address");
				}
				else 
				{
					reg_user.auditor_address = itr->value.GetString();
					util::formatAddress(reg_user.auditor_address);
				}

				itr = doc.FindMember("roles");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的 roles");
				}
				else 
				{
					// array: ["admin","nodeManage"]
					if(itr->value.IsArray() && !itr->value.Empty())
					{
						reg_user.roles.clear();
						SizeType size = itr->value.Size();
						for(SizeType i = 0; i < size; i++)
						{
							string str = itr->value[i].GetString();
							reg_user.roles.push_back(str);
						}

					//	bcwasm::println("reg_user.roles:", reg_user.roles);
					}

				//	reg_user.roles = itr->value.GetString();
				}

				return true;
			}
			// 解析返回code,msg,data参数
			bool desResultInfo(string& strResult, ResultInfo& res)
			{
				Document doc;
				doc.Parse<kParseDefaultFlags>(strResult.c_str());
				Document::AllocatorType &allocator = doc.GetAllocator();

				if (doc.HasParseError()) 
				{
					ParseErrorCode code = doc.GetParseError();
					bcwasm::println("解析返回信息失败，错误码：", code);
					return false;
				}

				Value::MemberIterator itr = doc.FindMember("code");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的code");
					return false;
				}
				else 
				{
					res.code = to_string(itr->value.GetInt());
				}
			/*
				itr = doc.FindMember("msg");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的msg");
				}
				else 
				{
					res.msg = itr->value.GetString();
				}

				itr = doc.FindMember("data");
				if (itr == doc.MemberEnd()) 
				{
					bcwasm::println("没找到对应的data");
				}
				else 
				{
				//	res.data = itr->value.GetString();
				}*/
				return true;
			}

			int storeRegisterRecord(struct RegisterInfo & reg_user, const string& key) 
			{
				std::string db_reg_record;
				serializeDbRegisterRecord(reg_user, db_reg_record);
				bcwasm::setState<std::string, std::string>(key, db_reg_record);

				return 0;
			}

			// 判断申请人/审核人是否是管理员角色
			bool isAdminRole(const string& address) const 
			{
				DeployedContract a(registAddr);
                string strRoleMageAddr = a.callString("getContractAddress", "__sys_RoleManager", "latest");

				DeployedContract role_c(strRoleMageAddr);
				/*
				bcwasm::byte ret = role_c.callInt64("isAdmin", address);
				if (ret != 0) 
				{
				//	bcwasm::println("调用者权限不合法，合约执行结果失败");
					return false;
				}*/
				string strResult = role_c.callString("getRolesByAddress", address.c_str());
				
				if(-1 == strResult.find("Admin") && -1 == strResult.find("chainCreator"))
				{
					bcwasm::println("调用者权限不合法，合约执行结果失败");
					return false;
				}
				return true;
			}

			// 判断用户/申请人/审核人是否是有效用户
			bool isValidUser(const string& userAddr) const
			{
				// 获取平台用户管理合约UserManage地址
				DeployedContract a(registAddr);
                string strUserMageAddr = a.callString("getContractAddress", "__sys_UserManager", "latest");
				// 调用UserManage合约isValidUser接口判断用户是否有效
				DeployedContract b(strUserMageAddr.c_str());
				int nRet = b.callInt64("isValidUser", userAddr);
				if(1 != nRet)
				{
					return false;
				}
				return true;
			}

			// 判断用户是否已经申请过
			bool isRegistered(std::string address) const 
			{
				util::formatAddress(address);
				std::string db_reg_record;
				getRegisterRecord(address, db_reg_record);

				if (db_reg_record.empty()) 
				{
					bcwasm::println("user:[", address, "] not found!");
					return false;
				}
				return true;
			}

			// 调用RoleManager合约setRole接口设置用户角色
			int setUserRole(const string userName, const string userAddr, const string& roles)
			{
				DeployedContract a(registAddr);
				string strRoleMageAddr = a.callString("getContractAddress", "__sys_RoleManager", "latest");
				DeployedContract RM_Contract(strRoleMageAddr.c_str());

				// 往角色管理合约写入角色信息
				string addr = userAddr;
				util::formatAddress(addr);
				int nRet = RM_Contract.callInt64("addRole", userName.c_str(), addr.c_str(), roles.c_str());
				return nRet;
			}	

			// 获取已经申请过的记录
			void getRegisterRecord(const std::string& address, std::string& db_reg_record) const 
			{
				string addr = address;
				util::formatAddress(addr);
				bcwasm::getState<std::string, std::string>(addr, db_reg_record);
			}

			// 处理返回结果
			const char* getResJson(const string& code, const string& message, const string& dataInfo) const
			{
				string strRetJson = "{\"code\":" + code + ",";
				strRetJson += "\"msg\":\"" + message + "\",";
				strRetJson += "\"data\":" + dataInfo + "}";
				return strRetJson.c_str();
			}

        private:
			ListAudit m_listAudit;
			ListActive m_listActive;
			ListReject m_listReject;
    };
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::UserRegister, registerUser)
BCWASM_ABI(SystemContract::UserRegister, approve)
BCWASM_ABI(SystemContract::UserRegister, getAccountByAddress)
BCWASM_ABI(SystemContract::UserRegister, getAccountByUsername)
BCWASM_ABI(SystemContract::UserRegister, getAccountsByStatus)
BCWASM_ABI(SystemContract::UserRegister, getStatusByAddress)
//bcwasm autogen begin
extern "C" { 
int registerUser(const char * registJson) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.registerUser(registJson);
}
int approve(const char * userAddress,int auditStatus) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.approve(userAddress,auditStatus);
}
const char * getAccountByAddress(const char * address) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.getAccountByAddress(address);
}
const char * getAccountByUsername(const char * UserName) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.getAccountByUsername(UserName);
}
const char * getAccountsByStatus(int pageNum,int pageSize,int accountStatus) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.getAccountsByStatus(pageNum,pageSize,accountStatus);
}
int getStatusByAddress(const char * address) {
SystemContract::UserRegister UserRegister_bcwasm;
return UserRegister_bcwasm.getStatusByAddress(address);
}
void init() {
SystemContract::UserRegister UserRegister_bcwasm;
UserRegister_bcwasm.init();
}

}
//bcwasm autogen end