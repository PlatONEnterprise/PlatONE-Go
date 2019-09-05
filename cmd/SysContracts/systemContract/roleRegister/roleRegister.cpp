#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include <vector>
#include "rapidjson/document.h"
#include "rapidjson/writer.h"
#include "rapidjson/stringbuffer.h"
#include <math.h>
#include "../common/util.hpp"

using namespace rapidjson;

namespace SystemContract
{
    const char *regManagerAddr = "0x0000000000000000000000000000000000000011";

    typedef struct RegisterUserInfo {
        std::string userAddress;       //账户地址
        std::string userName;            //用户名
        int roleRequireStatus;  //角色申请状态 1 申请中 2 已批准 3 已拒绝
        std::vector<std::string> requireRoles;
        std::string approver;         //审核人的地址
        BCWASM_SERIALIZE(RegisterUserInfo, (userAddress)(userName)(roleRequireStatus)(requireRoles)(approver));
    }RegisterUserInfo_st;
    
    char AddressInfoMapName[] = "AddressInfoMapName";
    typedef bcwasm::db::Map<AddressInfoMapName, std::string, RegisterUserInfo> AddressInfoMap_t;

    enum RegisterCode
    {
        R_SUCCESS = 0,
        R_ADDRRESS_STATUS_ERROR,
        R_INTERNAL_ERROR,
        R_REJECTED_ERROR,
        R_INPUT_FORMAT_ERROR,
        R_ROLE_ERROR,
        R_NO_REGISTER_INFO_ERROR,
        R_INVAID_ROLE_ERROR
    };

    enum ApproveCode 
    {
        A_SUCCESS = 0,
        A_ADDRRESS_FORMAT_ERROR,
        A_ROLE_ERROR,
        A_NO_REGISTER_INFO_ERROR,
        A_INTERNAL_ERROR,
        A_INVAID_APPROVE_CODE_ERROR,
        A_APPROVE_STATUS_ERROR,
        A_ADD_ROLE_ERROR,
    };

    enum Approve
    {
        S_APPROVING = 1,
        S_APPROVED,
        S_REJECTED,

    };

    class RoleRegister : public bcwasm::Contract
    {
        public:
            RoleRegister() {}

            void init() {
                util::registerContractFromInit("__sys_RoleRegister","1.0.0.0");
                bcwasm::println("init RoleRegister success...");
            }

            /// 定义Event.
            BCWASM_EVENT(Notify, uint64_t, const char *)

        public:
            // 注册角色
            // @roles required, eg: ["chainAdmin","nodeAdmin"]
            // @return 含义如下：
            //  0 注册成功
            //  1 地址状态不正确
            //  2 内部错误
            //  3 该地址已含被拒绝记录
            //  4 输入参数不正确，不是list类型
            //  5 调用者无权限调用本方法
            //  6 没有有效的申请信息
            //  7 角色名称无效
            int registerRole(const char* roles) {

                //检查地址状态
                std::string userAddrStr = bcwasm::origin().toString();
                util::formatAddress(userAddrStr);
                int userStatus = util::getUserStatus();
                if (0 != userStatus) {
                    string err = "ERR: [RoleRegister] [registerRole] User status is invalid. " + userAddrStr;
                    BCWASM_EMIT_EVENT(Notify, R_ADDRRESS_STATUS_ERROR, err.c_str());
                    bcwasm::println(err);
                    return R_ADDRRESS_STATUS_ERROR; // 1: 地址状态不正确
                }

                //0307: 如果已被拒绝，可再次申请
                /*
                std::map<std::string, RegisterUserInfo> addrMap = AddressInfoMap.get(); //临时变量
                std::map<std::string, RegisterUserInfo>::iterator regIt = addrMap.find(userAddress.toString());
                if(addrMap.end() != regIt) {
                    int roleRequireStatus = regIt->second.roleRequireStatus;        
                    if (3 == roleRequireStatus) 
                        return 3; //该地址已含被拒绝记录
                }
                */

                std::vector<std::string> rolestoRegister;  
                Document doc;
                std::string refinedRoles = util::removeSpace(roles);
                doc.Parse<0>(refinedRoles.c_str()); 
                if(!doc.IsArray()) {
                    string err = "ERR: [RoleRegister] [registerRole] input role array format is invalid. ";
                    BCWASM_EMIT_EVENT(Notify, R_INPUT_FORMAT_ERROR, err.c_str());
                    bcwasm::println(err);
                    return R_INPUT_FORMAT_ERROR; //输入参数不正确，不是list类型
                }

                for (Value::ConstValueIterator it = doc.Begin(); it != doc.End(); ++it) {
                    rolestoRegister.push_back(it->GetString());
                }
                
                std::vector<std::string> rolesOwned;
                util::getRoles(rolesOwned);

                //检查待申请的roles: 1.筛除不合法的role名称； 2.筛除已有role
                std::vector<std::string> requireRoles;
                for (auto const& it : rolestoRegister) {                     
                    std::string role = it;
                    if(!isValidRole(role) || role == gdef::chainCreator) {
                        string err = "ERR: [RoleRegister] [registerRole] not valid role name, and cannot register chainCreator.";
                        BCWASM_EMIT_EVENT(Notify, R_INVAID_ROLE_ERROR, err.c_str());
                        bcwasm::println(err);
                        return R_INVAID_ROLE_ERROR;
                    }
                    std::vector<std::string>::iterator itr = find(rolesOwned.begin(), rolesOwned.end(), role);
                    if (itr != rolesOwned.end())
                        continue;
                    requireRoles.push_back(role);   
                }

                if (requireRoles.size() <= 0) {
                    string err = "ERR: [RoleRegister] [registerRole] no valid role to register, please check if the roles are already owned or incorrect.";
                    BCWASM_EMIT_EVENT(Notify, R_NO_REGISTER_INFO_ERROR, err.c_str());
                    bcwasm::println(err);
                    return R_NO_REGISTER_INFO_ERROR; //没有有效的申请信息
                }
                
                std::string userName = getNamebyAddress(userAddrStr.c_str());

                RegisterUserInfo registerUserInfo = {
                    .userAddress = userAddrStr,
                    .userName = userName,
                    .roleRequireStatus = 1,
                    .requireRoles = requireRoles,
                    //.approver = std::string("0x11"),
                };

                AddressInfoMap.insert(userAddrStr, registerUserInfo); 

                string ok = "OK: [RoleRegister] [registerRole] Register success.";
                BCWASM_EMIT_EVENT(Notify, R_SUCCESS, ok.c_str());
                bcwasm::println(ok);
                return R_SUCCESS;
            }

            // 审批角色
            // @address required, eg: "0x0000000000000000000000000000000000000011"
            // @status required, eg: 2
            // @return 含义如下：
            //  0 注册成功
            //  1 账号地址格式错误
            //  2 角色权限不足
            //  3 没有角色申请信息 
            //  4 内部错误
            //  5 不是合法的审批值
            //  6 当前是非审批状态 
            
            int approveRole(const char* address, int status) {

                std::string addStr(address);
                util::formatAddress(addStr);

                if(!util::checkAddressFormat(addStr.c_str())) {
                    string err = "ERR: [RoleRegister] [approve] Address format is invalid. " + string(address);
                    BCWASM_EMIT_EVENT(Notify, A_ADDRRESS_FORMAT_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_ADDRRESS_FORMAT_ERROR;
                }
                
                std::string approverAddr = bcwasm::origin().toString();
                util::formatAddress(approverAddr);
                std::vector<std::string> regRoles;
                getRolesbyAddress(addStr.c_str(), regRoles);

                std::vector<std::string> approverRoles;
                util::getRoles(approverRoles);
                
                //检查是否有权限审核
                if (!util::ifHavePermission(approverRoles, regRoles)){
                    string err = "ERR: [RoleRegister] [approve] Approver has no proper roles to approve this address. " + addStr;
                    BCWASM_EMIT_EVENT(Notify, A_ROLE_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_ROLE_ERROR; //角色权限不足；
                }
                
                //检查地址状态
                RegisterUserInfo *userInfoPtr = AddressInfoMap.find(addStr);
                if(nullptr == userInfoPtr) {
                    string err = "ERR: [RoleRegister] [approve] This address has no roles to register. " + addStr;
                    BCWASM_EMIT_EVENT(Notify, A_NO_REGISTER_INFO_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_NO_REGISTER_INFO_ERROR; //没有角色申请信息 
                }

                int statusOwned = userInfoPtr->roleRequireStatus;

                if(S_APPROVING != statusOwned) {
                    string err = "ERR: [RoleRegister] [approve] the address is not waiting for approving now.";
                    BCWASM_EMIT_EVENT(Notify, A_APPROVE_STATUS_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_APPROVE_STATUS_ERROR; //当前是非审批状态
                }

                if (S_REJECTED == status) {

                    userInfoPtr->roleRequireStatus = S_REJECTED;
                    userInfoPtr->approver = approverAddr;
                    AddressInfoMap.update(addStr, *userInfoPtr);

                    string ok = "OK: [userRoleManger] [addRole] Registration rejected. ";
                    BCWASM_EMIT_EVENT(Notify, A_SUCCESS, ok.c_str());
                    bcwasm::println(ok);
                    return A_SUCCESS;
                }
                else if (S_APPROVED == status) {
                    std::vector<std::string> requireRoles = userInfoPtr->requireRoles;
                    Document doc;
                    doc.SetObject();
                    Document::AllocatorType &allocator=doc.GetAllocator();
                    Value roles(kArrayType);

                    for (auto const& it : requireRoles) {                       
                        std::string role(it);
                        Value v(role.c_str(), allocator);
                        roles.PushBack(v, allocator);
                    }
                    std::string roleString = jsonToString(roles);
                    std::string userName = (*(AddressInfoMap.find(addStr))).userName;
                        
                    //调用角色管理合约
                    bcwasm::DeployedContract regManagerContract(regManagerAddr);
                    std::string strRoleManagerAddr = regManagerContract.callString("getContractAddress", "__sys_RoleManager", "latest");

                    if ("" == strRoleManagerAddr) {
                        string err = "ERR: [RoleRegister] [approve] no roleManage Contract.";
                        BCWASM_EMIT_EVENT(Notify, A_INTERNAL_ERROR, err.c_str());
                        bcwasm::println(err);
                        return A_INTERNAL_ERROR; //未找到角色管理合约
                    }
                    bcwasm::Address roleManagerContractAddress(strRoleManagerAddr.c_str());
                    bcwasm::DeployedContract rmc(roleManagerContractAddress);
                        
                    int ret = rmc.callInt64("addRole", userName.c_str(), addStr.c_str(), roleString.c_str());
                    if (0 != ret) {
                        string err = "ERR: [RoleRegister] [approve] addRole error. " + std::to_string(ret);
                        BCWASM_EMIT_EVENT(Notify, A_ADD_ROLE_ERROR, err.c_str());
                        bcwasm::println(err);
                        return A_ADD_ROLE_ERROR; //未找到角色管理合约
                    }

                    userInfoPtr->roleRequireStatus = S_APPROVED;
                    userInfoPtr->approver = approverAddr;
                    AddressInfoMap.update(addStr, *userInfoPtr);

                    string ok = "OK: [RoleRegister] [approve] Registration approved. ";
                    BCWASM_EMIT_EVENT(Notify, A_SUCCESS, ok.c_str());
                    bcwasm::println(ok);
                    return A_SUCCESS;
                }
                else {
                    string err = "ERR: [RoleRegister] [approve] not a valid approve code.";
                    BCWASM_EMIT_EVENT(Notify, A_INVAID_APPROVE_CODE_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_INVAID_APPROVE_CODE_ERROR; //不是合法的审批值
                }   
            }
            
            const char * getRegisterInfoByAddress(const char* address) const{

                std::string addStr(address);
                util::formatAddress(addStr);
                
                Document doc;
                doc.SetObject();
                Document::AllocatorType &allocator=doc.GetAllocator();
                int code = 0;
                std::string msg = "";
                Value dataObject(kObjectType);

                const RegisterUserInfo *userInfoPtr = AddressInfoMap.find(addStr);
                if(nullptr == userInfoPtr) { 
                    code = 1;
                    msg = "not found";
                    dataObject.SetString("");
                }
                else {
                    //构造dataObject
                    Value keyUserAddress("userAddress", allocator);
                    Value keyUserName("userName", allocator);
                    Value keyStatus("roleRequireStatus", allocator);
                    Value keyRequireRoles("requireRoles", allocator);
                    Value keyAproveor("approver", allocator);

                    Value valName(userInfoPtr->userName.c_str(), allocator);
                    Value valAddress(userInfoPtr->userAddress.c_str(), allocator);
                    Value valAproveor(userInfoPtr->approver.c_str(), allocator);

                    dataObject.AddMember(keyUserAddress, valAddress, allocator);
                    dataObject.AddMember(keyUserName, valName, allocator);
                    dataObject.AddMember(keyStatus, userInfoPtr->roleRequireStatus, allocator);
                    vector<std::string> roleVector = (*(AddressInfoMap.find(addStr))).requireRoles;
                    Value roles(kArrayType);
                    for (auto it = roleVector.begin();it != roleVector.end(); ++it) {                                              
                        Value v(it->c_str(), allocator);
                        roles.PushBack(v, allocator);
                    }
                    dataObject.AddMember(keyRequireRoles, roles, allocator);
                    dataObject.AddMember(keyAproveor, valAproveor, allocator);

                    code = 0;
                    msg = "ok";
                }

                doc.AddMember("code", code, allocator);
                doc.AddMember("msg", StringRef(msg.c_str()), allocator);
                doc.AddMember("data", dataObject, allocator);

                StringBuffer jsonBuffer;  
                Writer<StringBuffer> writer(jsonBuffer);
                doc.Accept(writer);  
                
                size_t size = jsonBuffer.GetSize();
                char* buf = (char *)malloc(size + 1);
                if (buf == nullptr) {
                    bcwasm::println("Malloc failed.");
                    return buf;
                }
                memset(buf, 0, size + 1);
                strcpy(buf, jsonBuffer.GetString());

                return buf;
            }

            const char * getRegisterInfoByName(const char* name) const{

                for (auto it = AddressInfoMap.begin(); it !=AddressInfoMap.end(); ++it) {
                    RegisterUserInfo userInfo = it->second();
                    if (name == userInfo.userName) {
                        return getRegisterInfoByAddress(it->first().c_str());
                    }
                }

                Document doc;
                doc.SetObject();
                Document::AllocatorType &allocator=doc.GetAllocator();

                int code = 1;
                std::string msg = "not found";
                Value dataObject(kObjectType);
                dataObject.SetString("");

                doc.AddMember("code", code, allocator);
                doc.AddMember("msg", StringRef(msg.c_str()), allocator);
                doc.AddMember("data", dataObject, allocator);

                StringBuffer jsonBuffer;  
                Writer<StringBuffer> writer(jsonBuffer);
                doc.Accept(writer);  
                
                size_t size = jsonBuffer.GetSize();
                char* buf = (char *)malloc(size + 1);
                if (buf == nullptr) {
                    bcwasm::println("Malloc failed.");
                    return buf;
                }
                memset(buf, 0, size + 1);
                strcpy(buf, jsonBuffer.GetString());

                return buf;
            }

            const char * getRegisterInfosByStatus(int status, int pageNum, int pageSize) const {
                Document doc;
                doc.SetObject();
                Document::AllocatorType &allocator=doc.GetAllocator();
                int code = 1;
                std::string msg = "not found";
                Value infoList(kArrayType);
                
                std::vector<RegisterUserInfo> userInfo;

                for (auto it = AddressInfoMap.begin(); it !=AddressInfoMap.end(); ++it) {
                    RegisterUserInfo info = it->second();
                    if (status == info.roleRequireStatus) {
                        userInfo.push_back(info);
                    }
                }

		        if(!userInfo.empty()) {
                    int count = 0;
                    int begin = pageNum * pageSize;
                    int end = (pageNum + 1) * pageSize;

                    for (auto const& it : userInfo) {                       
                        if(count  < begin  || count > end) {
                            count ++;
                            continue;
                        }
                        const char* info = getRegisterInfoByAddress(it.userAddress.c_str());
                        Document d;
                        d.Parse<0>(info);
                        infoList.PushBack(d["data"], allocator);
                        count ++;
                    }
                    code = 0;
                    msg = "ok";
		        }

                doc.AddMember("code", code, allocator);
                doc.AddMember("msg", StringRef(msg.c_str()), allocator);
                doc.AddMember("data", infoList, allocator);
                return util::documentToString(&doc);                
            }

        private:

            bool isValidRole(const std::string &role) {
                auto itr = validRoles.find(role);
                if(itr != validRoles.end())
                    return true;
                else return false;
            }

            int getRoleLevel(bcwasm::Address address) {
                std::string addStr(address.toString());
                util::formatAddress(addStr);

                std::vector<std::string> roles;
                getRolesbyAddress( addStr.c_str(), roles);
                int roleLevel = 3;
                //for (std::vector<std::string>::iterator it = roles.begin(); it != roles.end(); ++it) {
                for (auto const& it : roles) { 
                    int roleLevel_itr = validRoles.find(it)->second;
                    roleLevel = roleLevel_itr < roleLevel ? roleLevel_itr : roleLevel;
                }
                return roleLevel;
            }

            void getRolesbyAddress(const char *userAddress, std::vector<std::string> &roles) {

                std::string addStr(userAddress);
                util::formatAddress(addStr);

                bcwasm::DeployedContract regManagerContract(regManagerAddr);
                std::string strRoleManagerAddr = regManagerContract.callString("getContractAddress", "__sys_RoleManager", "latest");

                if ("" == strRoleManagerAddr) {
                    return;
                }

                bcwasm::Address rmAddr(strRoleManagerAddr);
                bcwasm::DeployedContract roleManagerContract(rmAddr);
                std::string strRoles = roleManagerContract.callString("getRolesByAddress", addStr.c_str());
                if ("" == strRoles) {
                    return;
                }

                Document doc;
                doc.Parse<0>(strRoles.c_str());
                if(doc.HasMember("code") && 0 != doc["code"]) {
                    bcwasm::println("[RoleRegister] invalid roles.");
                    return;
                }
                if(!doc["data"].IsArray()) {
                    bcwasm::println("[RoleRegister] invaid roles format.");
                    return;
                }
                for(Value::ConstValueIterator itr = doc["data"].Begin(); itr != doc["data"].End(); ++itr) {
                    roles.push_back(itr->GetString());
                }
                return;
            }

            std::string getNamebyAddress(const char *userAddress) {

                std::string addStr(userAddress);
                util::formatAddress(addStr);

                bcwasm::DeployedContract regManagerContract(regManagerAddr);
                std::string strUserManagerAddr = regManagerContract.callString("getContractAddress", "__sys_UserManager", "latest");

                if ("" == strUserManagerAddr) {
                    return "";
                }

                bcwasm::Address umAddr(strUserManagerAddr);
                bcwasm::DeployedContract userManagerContract(umAddr);
                std::string account = userManagerContract.callString("getAccountByAddress", addStr.c_str());
                Document doc;
                doc.Parse<0>(account.c_str());

                bcwasm::println("[RoleRegister] account code: ", to_string(doc["code"].GetInt()));
                if(doc.HasMember("code") && 0 != doc["code"].GetInt()) {
                    bcwasm::println("[RoleRegister] invalid account info.");
                    return "";
                }
                if(!doc["data"].HasMember("name")) {
                    bcwasm::println("[RoleRegister] invaid name format.");
                    return "";
                }

                std::string name = doc["data"]["name"].GetString();
                return name;
            }

            std::string jsonToString(const Value& valObj) {
                
                StringBuffer sbBuf;
                Writer<StringBuffer> jWriter(sbBuf);
                valObj.Accept(jWriter);
                return std::string(sbBuf.GetString());
            }

        private:

            AddressInfoMap_t AddressInfoMap;

	    //可用角色列表: value 1-->3，审批等级由高到低, 1最高, 3最低。
            std::map<std::string, int > validRoles = {
                {"chainCreator", 1},
                {"chainAdmin",2},
                {"nodeAdmin",3},
                {"contractAdmin",3},
                {"contractDeployer",3},
                {"contractCaller",3}
            };
    };
} // namespace RoleRegister


BCWASM_ABI(SystemContract::RoleRegister, registerRole)
BCWASM_ABI(SystemContract::RoleRegister, approveRole)
BCWASM_ABI(SystemContract::RoleRegister, getRegisterInfoByAddress)
BCWASM_ABI(SystemContract::RoleRegister, getRegisterInfoByName)
BCWASM_ABI(SystemContract::RoleRegister, getRegisterInfosByStatus)
//bcwasm autogen begin
extern "C" { 
int registerRole(const char * roles) {
SystemContract::RoleRegister RoleRegister_bcwasm;
return RoleRegister_bcwasm.registerRole(roles);
}
int approveRole(const char * address,int status) {
SystemContract::RoleRegister RoleRegister_bcwasm;
return RoleRegister_bcwasm.approveRole(address,status);
}
const char * getRegisterInfoByAddress(const char * address) {
SystemContract::RoleRegister RoleRegister_bcwasm;
return RoleRegister_bcwasm.getRegisterInfoByAddress(address);
}
const char * getRegisterInfoByName(const char * name) {
SystemContract::RoleRegister RoleRegister_bcwasm;
return RoleRegister_bcwasm.getRegisterInfoByName(name);
}
const char * getRegisterInfosByStatus(int status,int pageNum,int pageSize) {
SystemContract::RoleRegister RoleRegister_bcwasm;
return RoleRegister_bcwasm.getRegisterInfosByStatus(status,pageNum,pageSize);
}
void init() {
SystemContract::RoleRegister RoleRegister_bcwasm;
RoleRegister_bcwasm.init();
}

}
//bcwasm autogen end