#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include <chrono>
#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include "../common/util.hpp"

using namespace rapidjson;
using namespace std;

namespace SystemContract
{
    struct userRole{
        bcwasm::Address address;
        string name;
        vector<string> roles;
        BCWASM_SERIALIZE(userRole,(address)(name)(roles));
    };

    char mapName[] = "mapName";

//    typedef bcwasm::StorageType<mapName,map<string,userRole>> pMap;
    typedef bcwasm::db::Map<mapName, string, userRole> pMap;

    class RoleManager : public bcwasm::Contract
    {
    private:
        pMap mapUserRoles;
        bcwasm::Address regManagerAddr;

        public:
            RoleManager() {}
            void init()
            {
                //初始化一个root用户，也就是超管
                std::string origin = bcwasm::origin().toString();
                util::formatAddress(origin);

                userRole urole;
                urole.name = "root";
                urole.address = bcwasm::Address(origin);
                urole.roles.push_back(util::CHAIN_CREATOR);
                mapUserRoles.insert(origin, urole);

                util::registerContractFromInit("__sys_RoleManager","1.0.0.0");
                bcwasm::println("init RoleManager contract success");
            }

        /// 定义Event
        BCWASM_EVENT(Notify, uint64_t, const char *)

       public:
        enum RoleAddCode
        {
            A_SUCCESS = 0,
            A_ADDRESS_FORMAT_ERROR,
            A_CALLER_STATUS_ERROR,
            A_USER_STATUS_ERROR,
            A_INVALID_ROLE_ERROR,
            A_NO_PERMISSION,
            A_NAME_ERROR,
            A_ADD_ROLE_ERROR
        };
      public:
        enum  RoleRemoveCode
        {
            R_SUCCESS = 0,
            R_ADDRESS_FORMAT_ERROR,
            R_CALLER_STATUS_ERROR,
            R_NO_ADDRESS,
            R_INVALID_ROLE_ERROR,
            R_NO_PERMISSION
        };



        public:
        int addRole(const char *name, const char *address, const char *roles)
            {
                //TODO check name if invalid?
                std::string addr = address;
                util::formatAddress(addr);

                if (!util::checkAddressFormat(addr.c_str()))
                {
                    string err = "ERR: [RoleManager] [addRole] Address format is invalid. " + string(address);
                    BCWASM_EMIT_EVENT(Notify, A_ADDRESS_FORMAT_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_ADDRESS_FORMAT_ERROR;
                }

                // check the status of caller
                 int userStatus = util::getUserStatus();
                 if (0 != userStatus)
                 {
                     string err = "ERR: [RoleManager]  [addRole] Caller unavailable, status: " + to_string(userStatus);
                     BCWASM_EMIT_EVENT(Notify, A_CALLER_STATUS_ERROR, err.c_str());
                     bcwasm::println(err);
                    
                     return A_CALLER_STATUS_ERROR;
                 }

                // check the status of user who being added role
                 userStatus = util::getUserStatusByUserAddr(string(address));
                 if (0 != userStatus)
                 {
                     string err = "ERR: [RoleManager]  [addRole] User unavailable, status: " + to_string(userStatus) + " ,  Address: " + string(address);
                     BCWASM_EMIT_EVENT(Notify, A_USER_STATUS_ERROR, err.c_str());
                     bcwasm::println(err);
                     return A_USER_STATUS_ERROR;
                 }

                Document doc;
                doc.Parse(roles);

                userRole urole;
                urole.name = name;
                urole.address = bcwasm::Address(addr);

                for (Value::ConstValueIterator itr = doc.Begin(); itr != doc.End(); ++itr)
                {
                    urole.roles.push_back(itr->GetString());
                }

                if ( !util::isRolesValid(urole.roles) )
                {
                    string err = "ERR: [RoleManager] [addRole] Roles invalid: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, A_INVALID_ROLE_ERROR, err.c_str());
                    bcwasm::println(err);
                    return A_INVALID_ROLE_ERROR;
                }

                //判断权限
                string callerAddrss = bcwasm::origin().toString();
                util::formatAddress(callerAddrss);

                const userRole *uRolePtr = mapUserRoles.find(callerAddrss);
                if(nullptr == uRolePtr) {
                    string err = "ERR: [RoleManager] [removeRole] No permission to remove for roles: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, A_NO_PERMISSION, err.c_str());
                    bcwasm::println(err);

                    return A_NO_PERMISSION;
                }

                // auto v = (*mapUserRoles).find(callerAddrss);
                // if ((*mapUserRoles).end() == v){
                //     string err = "ERR: [RoleManager] [removeRole] No permission to remove for roles: " + string(roles);
                //     BCWASM_EMIT_EVENT(Notify, 5, err.c_str());
                //     bcwasm::println(err);

                //     return 5;
                // }

                vector<string> callerRoles = (*(mapUserRoles.find(callerAddrss))).roles;
                if ( !util::ifHavePermission(callerRoles,  urole.roles))
                {
                    string err = "ERR: [RoleManager] [addRole] No permission to addrole for roles: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, A_NO_PERMISSION, err.c_str());
                    bcwasm::println(err);
                    return A_NO_PERMISSION;
                }

                userRole tmpUserRole;
                int ret = doGetUserRoleInfoByName(name,tmpUserRole);
                if (0 == ret)//用户名已经存在
                {
                    if ( string("0x")+tmpUserRole.address.toString() != addr)
                    {
                        string err = "ERR: [RoleManager] [addRole] Name must be unique " + string(name);
                        BCWASM_EMIT_EVENT(Notify, A_NAME_ERROR, err.c_str());
                         bcwasm::println(err);
                        return  A_NAME_ERROR;// 用户名必须唯一
                    }
                }

                const userRole *uRolePtr2 = mapUserRoles.find(addr);
                if(nullptr == uRolePtr2) {
                    mapUserRoles.insert(addr, urole);                   
                    string ok = "OK: [RoleManager] [addRole] Add roles success. " + string(addr) + string(roles);
                    BCWASM_EMIT_EVENT(Notify, A_SUCCESS, ok.c_str());
                    bcwasm::println(ok);
                    
                    return A_SUCCESS;
                } else{
                    //存在并用户名一样，则更新
                    userRole urole2 = *(mapUserRoles.find(addr));
                    if(strcmp(urole2.name.c_str(), name) == 0)
                    {
                        for (Value::ConstValueIterator itr = doc.Begin(); itr != doc.End(); ++itr)
                        { 
                            bool isAlreadyHave = false;
                            for (auto const& r : urole2.roles)
                            {
                                if ( r == itr->GetString())
                                {
                                    isAlreadyHave = true;
                                    break;
                                }
                            }

                            //如果已经存在了此角色
                            if ( !isAlreadyHave )
                            {
                                urole2.roles.push_back(itr->GetString());
                            }
                        }

                        mapUserRoles.update(addr, urole2);
                        
                        string ok = "OK: [RoleManager] [addRole] Update roles success. address: " + string(addr) + " ,  roles: " + string(roles);
                        BCWASM_EMIT_EVENT(Notify, A_SUCCESS, ok.c_str());
                        bcwasm::println(ok);

                        return A_SUCCESS;
                    }
                }
                //新增
                // auto v2 = mapUserRoles.get().find(addr);
                // if (mapUserRoles.get().end() == v2){
                //     (*mapUserRoles)[addr] = urole;
                    
                //     string ok = "OK: [RoleManager] [addRole] Add roles success. " + string(addr) + string(roles);
                //     BCWASM_EMIT_EVENT(Notify, 0, ok.c_str());
                //     bcwasm::println(ok);
                    
                //     return 0;
                // } else{
                //     //存在并用户名一样，则更新
                //     userRole urole2 = (*mapUserRoles)[addr];
                //     if(strcmp(urole2.name.c_str(), name) == 0)
                //     {
                //         for (Value::ConstValueIterator itr = doc.Begin(); itr != doc.End(); ++itr)
                //         { 
                //             bool isAlreadyHave = false;
                //             for (auto const& r : urole2.roles)
                //             {
                //                 if ( r == itr->GetString())
                //                 {
                //                     isAlreadyHave = true;
                //                     break;
                //                 }
                //             }

                //             //如果已经存在了此角色
                //             if ( !isAlreadyHave )
                //             {
                //                 urole2.roles.push_back(itr->GetString());
                //             }
                //         }

                //         (*mapUserRoles)[addr] = urole2;
                        
                //         string ok = "OK: [RoleManager] [addRole] Update roles success. address: " + string(addr) + " ,  roles: " + string(roles);
                //         BCWASM_EMIT_EVENT(Notify, 0, ok.c_str());
                //         bcwasm::println(ok);

                //         return 0;
                //     }
                // }

                string err = "ERR: [RoleManager] [addRole] Add Roles failed " + string(addr) + string(roles);
                BCWASM_EMIT_EVENT(Notify, A_ADD_ROLE_ERROR, err.c_str());
                bcwasm::println(err);

                return A_ADD_ROLE_ERROR;
            }

            int removeRole(const char *address,const char *roles)
            {
                std::string addr = address;
                util::formatAddress(addr);

                if (!util::checkAddressFormat(addr.c_str()))
                {
                    string err = "ERR: [RoleManager] [addRole] Address format is invalid. " + string(address);
                    BCWASM_EMIT_EVENT(Notify, R_ADDRESS_FORMAT_ERROR, err.c_str());
                    bcwasm::println(err);
                    return R_ADDRESS_FORMAT_ERROR;
                }

                 int userStatus = util::getUserStatus();
                 if (0 != userStatus)
                 {
                     string err = "ERR: [RoleManager]  [removeRole] Caller unavailable, status: " + to_string(userStatus) + " , Address: " + string(address);
                     BCWASM_EMIT_EVENT(Notify, R_CALLER_STATUS_ERROR, err.c_str());
                     bcwasm::println(err);
                     return R_CALLER_STATUS_ERROR;
                 }

                const userRole *uRolePtr = mapUserRoles.find(addr);
                if (nullptr == uRolePtr) {
                    string err = "ERR: [RoleManager]  [removeRole] Cannot found address: " +  string(address);
                    BCWASM_EMIT_EVENT(Notify, R_NO_ADDRESS, err.c_str());
                    bcwasm::println(err);

                    return R_NO_ADDRESS;
                }
                // auto v = (*mapUserRoles).find(addr);
                // if ((*mapUserRoles).end() == v){
                //     string err = "ERR: [RoleManager]  [removeRole] Cannot found address: " +  string(address);
                //     BCWASM_EMIT_EVENT(Notify, 3, err.c_str());
                //     bcwasm::println(err);

                //     return 3;
                // }

                userRole urole = *(mapUserRoles.find(addr));

                Document doc;
                doc.Parse(roles);

                //判断权限
                vector<string> rls;
                for (Value::ConstValueIterator itr = doc.Begin(); itr != doc.End(); ++itr)
                {
                    rls.push_back(itr->GetString());
                }
                 
                if ( !util::isRolesValid(rls) )
                {
                    string err = "ERR: [RoleManager] [addRole] Roles invalid: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, R_INVALID_ROLE_ERROR, err.c_str());
                    bcwasm::println(err);

                    return R_INVALID_ROLE_ERROR;
                }

                //check if have permission
                string callerAddrss = bcwasm::origin().toString();
                util::formatAddress(callerAddrss);

                const userRole *uRolePtr2 = mapUserRoles.find(callerAddrss);
                if (nullptr == uRolePtr2) {
                    string err = "ERR: [RoleManager] [removeRole] No permission to remove for roles: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, R_NO_PERMISSION, err.c_str());
                    bcwasm::println(err);

                    return R_NO_PERMISSION;
                }
                // auto v2 = (*mapUserRoles).find(callerAddrss);
                // if ((*mapUserRoles).end() == v2){
                //     string err = "ERR: [RoleManager] [removeRole] No permission to remove for roles: " + string(roles);
                //     BCWASM_EMIT_EVENT(Notify, 5, err.c_str());
                //     bcwasm::println(err);

                //     return 5;
                // }

                vector<string> callerRoles = (*(mapUserRoles.find(callerAddrss))).roles;
                if ( !util::ifHavePermission(callerRoles, rls))
                {
                    string err = "ERR: [RoleManager] [removeRole] No permission to remove for roles: " + string(roles);
                    BCWASM_EMIT_EVENT(Notify, R_NO_PERMISSION, err.c_str());
                    bcwasm::println(err);

                    return R_NO_PERMISSION;
                }

                vector<string> svecRoles = urole.roles;
                for (Value::ConstValueIterator itr = doc.Begin(); itr != doc.End(); ++itr)
                {
                    for(vector<string>::iterator iter = svecRoles.begin(); iter != svecRoles.end(); )
                    {
                        string str(*iter);
                        if (itr->GetString() == str)
                        {
                            iter = svecRoles.erase(iter);
                        } else{
                            ++iter;
                        }
                    }
                }

                urole.roles = svecRoles;
                mapUserRoles.insert(addr, urole);

                string ok = "OK: [RoleManager] [removeRole] Remove roles success.  address: " + string(addr) +"  , roles: " + string(roles);
                BCWASM_EMIT_EVENT(Notify, R_SUCCESS, ok.c_str());
                bcwasm::println(ok);

                return R_SUCCESS;
            }

            const char* getRolesByAddress(const char *address)const
            {
                std::string addr = address;
                util::formatAddress(addr);

                if (!util::checkAddressFormat(addr.c_str()))
                {
                    string err = "Address format is invalid. " + string(address);
                    return util::errorResultToString(1, err);
                }

                const userRole *uRolePtr = mapUserRoles.find(addr);
                if (nullptr == uRolePtr) {
                    return util::errorResultToString(2, "Not Found");
                }
                // auto v = mapUserRoles.get().find(addr);
                // if (mapUserRoles.get().end() == v)
                // {
                //     return util::errorResultToString(2, "Not Found");
                // }

                userRole urole = *(mapUserRoles.find(addr));

                return  util::vectorResultToString(urole.roles);
            }

            const char* getRolesByName(const char *name)const
            {
                //TODO check name if invalid?
                vector<string> svec = doGetRolesByName(name);
                if (svec.empty())
                {
                    return util::errorResultToString(1,"Not Found");
                }

                return util::vectorResultToString(svec);
            }

            const char* getAccountsByRole(const char *role)const
            {
                if (!util::isRoleValid(role))
                {
                    return util::errorResultToString(1,"role parameter invalid");
                }

                Document doc;
                doc.SetObject();

                Document::AllocatorType& allocator = doc.GetAllocator();
                Value array(kArrayType);

                //const std::set<string>& keys = mapUserRoles.getKeys();

                for(auto iter = mapUserRoles.begin(); iter != mapUserRoles.end(); ++iter) {
                    userRole uRole = iter->second();
                    vector<string> roles = uRole.roles;
                    for(auto const& r : roles)
                    {
                        if ( r == role )
                        {
                           Value account(kObjectType);

                           Value valName(uRole.name.c_str(),allocator);
                           account.AddMember("name",valName,allocator);

                           Value valAddress( (string("0x")+uRole.address.toString()).c_str(),allocator);
                           account.AddMember("address",valAddress,allocator);

                            array.PushBack(account,allocator);
                            bcwasm::println(uRole.address.toString().c_str());
                            break;
                        }
                    }
                }


                // map<string,userRole> tmp = mapUserRoles.get();
                // map<string,userRole>::iterator it = tmp.begin();
                // while(it !=tmp.end())
                // {
                //     vector<string> roles = it->second.roles;
                //     for(auto const& r : roles)
                //     {
                //         if ( r == role )
                //         {
                //            Value account(kObjectType);

                //            Value valName(it->second.name.c_str(),allocator);
                //            account.AddMember("name",valName,allocator);

                //            Value valAddress( (string("0x")+it->second.address.toString()).c_str(),allocator);
                //            account.AddMember("address",valAddress,allocator);

                //             array.PushBack(account,allocator);
                //             bcwasm::println(it->second.address.toString().c_str());
                //             break;
                //         }
                //     }

                //     it++;
                // }

                if (array.Empty()){
                    return util::errorResultToString(2,"Not Found");
                }

                doc.AddMember("code",0,allocator);
                doc.AddMember("msg","Success",allocator);
                doc.AddMember("data",array,allocator);
                return util::documentToString(&doc);
            }

            int hasRole(const char* addr, const char* role) const
            {
                string strAddr = addr;
                util::formatAddress(strAddr);
                
                //const std::set<string>& keys = mapUserRoles.getKeys();
                if (mapUserRoles.size() == 0)
                {
                    return 0;
                }

                // if (mapUserRoles.get().count(strAddr) == 0)
                // {
                //     return 0;
                // }

                userRole urole = *(mapUserRoles.find(strAddr));

                for (auto const& r : urole.roles)
                {
                    if (strcasecmp(r.c_str(), role) == 0)
                        return 1;
                }

                return 0;
            }

        private:
            int doGetUserRoleInfoByName(const char *name, userRole& urole)const
            {
                //必须有这一步，把得到的map放到一个临时变量，不然就无法获取里面struct中的vector数据
                
                //const std::set<string>& keys = mapUserRoles.getKeys();

                for(auto iter = mapUserRoles.begin(); iter != mapUserRoles.end(); ++iter) {
                    userRole uRole = iter->second();
                    if (string(name) == uRole.name)
                    {
                        urole.roles.assign(uRole.roles.begin(),uRole.roles.end());
                        urole.name = uRole.name;
                        urole.address = uRole.address;
                        return 0;
                    }
                }
                
                
                // map<string,userRole> tmp = mapUserRoles.get();
                // map<string,userRole>::iterator it = tmp.begin();

                // while(it != tmp.end())
                // {
                //     if (name == it->second.name)
                //     {
                //         urole.roles.assign(it->second.roles.begin(),it->second.roles.end());
                //         urole.name = it->second.name;
                //         urole.address = it->second.address;

                //         return 0;
                //     }

                //     it++;
                // }

                return  1;//not found
            }

            vector<string> doGetRolesByName(const char *name)const
            {
                //必须有这一步，把得到的map放到一个临时变量，不然就无法获取里面struct中的vector数据
                //const std::set<string>& keys = mapUserRoles.getKeys();

                for(auto iter = mapUserRoles.begin(); iter != mapUserRoles.end(); ++iter) {
                    userRole uRole = iter->second();
                    if (string(name) == uRole.name)
                    {
                        return uRole.roles;
                    }
                }
          
                // map<string,userRole> tmp = mapUserRoles.get();
                // map<string,userRole>::iterator it = tmp.begin();
                
                // while(it != tmp.end())
                // {
                //     if (string(name) == it->second.name)
                //     {
                //        return it->second.roles;
                //     }

                //     it++;
                // }

                vector<string> svec;
                return  svec;
            }
    };
}

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::RoleManager, addRole)
BCWASM_ABI(SystemContract::RoleManager, removeRole)
BCWASM_ABI(SystemContract::RoleManager, getRolesByAddress)
BCWASM_ABI(SystemContract::RoleManager, getRolesByName)
BCWASM_ABI(SystemContract::RoleManager, getAccountsByRole)
BCWASM_ABI(SystemContract::RoleManager, hasRole)

//bcwasm autogen begin
extern "C" { 
int addRole(const char * name,const char * address,const char * roles) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.addRole(name,address,roles);
}
int removeRole(const char * address,const char * roles) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.removeRole(address,roles);
}
const char * getRolesByAddress(const char * address) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.getRolesByAddress(address);
}
const char * getRolesByName(const char * name) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.getRolesByName(name);
}
const char * getAccountsByRole(const char * role) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.getAccountsByRole(role);
}
int hasRole(const char * addr,const char * role) {
SystemContract::RoleManager RoleManager_bcwasm;
return RoleManager_bcwasm.hasRole(addr,role);
}
void init() {
SystemContract::RoleManager RoleManager_bcwasm;
RoleManager_bcwasm.init();
}

}
//bcwasm autogen end