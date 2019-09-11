#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include "gdef.hpp"

using namespace rapidjson;
using namespace std;

namespace util {
    const char *regManagerAddr = "0x0000000000000000000000000000000000000011";

    const char* CHAIN_CREATOR = "chainCreator";
    const char* CHAIN_ADMIN = "chainAdmin";
    const char* NODE_ADMIN = "nodeAdmin";
    const char* CONTRACT_ADMIN = "contractAdmin";
    const char* CONTRACT_DEPLOYER = "contractDeployer";

    const int CONTRACT_NAME_LENGTH_MIN = 2;
    const int CONTRACT_NAME_LENGTH_MAX = 1024;

    bool isSpace (unsigned char x) {
            return std::isspace(x); 
    }

    std::string removeSpace(const char* input){
        string inputStr(input);
        inputStr.erase(remove_if(inputStr.begin(),inputStr.end(), isSpace), inputStr.end());

        return inputStr;
    }

    int registerContract(string name, string version, bcwasm::Address addr) {
        bcwasm::DeployedContract regManagerContract(regManagerAddr);
        regManagerContract.call("cnsRegister", name.c_str(), version.c_str(), string("0x" + addr.toString()).c_str());

        return 0;
    }

    int registerContractFromInit(string name, string version) {
        bcwasm::DeployedContract regManagerContract(regManagerAddr);
        regManagerContract.call("cnsRegisterFromInit", name.c_str(), version.c_str());

        return 0;
    }

    void formatAddress(std::string& addr) {
        if (addr.find("0x") != 0)
            addr = std::string("0x") + addr;
        std::transform(addr.begin(), addr.end(), addr.begin(), ::tolower);
    }

    bool checkAddressFormat(const char *addr)
    {
        if (addr == nullptr)
            return false;
        if (strlen(addr) == 0)
            return false;

        std::string address = addr;
        // must start with "0x"
        if (address.find("0x") != 0)
            return false;
        if (address.length() != 42)
            return false;

        for (int i = 2; i < address.length(); i++)
        {
            if (!isxdigit(address[i]))
                return false;
        }

        return true;
    }

    const  char* documentToString(Document* doc)
    {
        StringBuffer jsonBuffer;
        Writer<StringBuffer> writer(jsonBuffer);
        doc->Accept(writer);

        size_t size = jsonBuffer.GetSize();
        char* buf = (char *)malloc(size + 1);
        memset(buf, 0, size + 1);
        strcpy(buf, jsonBuffer.GetString());

        return buf;
    }

    const char* errorResultToString(int code,string msg)
    {
        Document doc;
        doc.SetObject();

        Document::AllocatorType& allocator = doc.GetAllocator();
        doc.AddMember("code",code,allocator);
        doc.AddMember("msg",StringRef(msg.c_str()),allocator);
        doc.AddMember("data","",allocator);

        return documentToString(&doc);
    }

    const char* successResultToString(Document *doc, int code,string msg,Value* data)
    {
        Document::AllocatorType& allocator = doc->GetAllocator();
        doc->AddMember("code",code,allocator);
        doc->AddMember("msg",StringRef(msg.c_str()),allocator);
        doc->AddMember("data",*data,allocator);

        return documentToString(doc);
    }

    const char* vectorResultToString(vector<string> vec)
    {
        Document doc;
        doc.SetObject();

        Document::AllocatorType& allocator = doc.GetAllocator();
        Value array(kArrayType);

        for(vector<string>::iterator iter = vec.begin();iter != vec.end();iter++){
            string str(*iter);
            Value v(str.c_str(),allocator);
            array.PushBack(v,allocator);
        }

        return successResultToString(&doc, 0, "Success",&array);
    }

    //参数说明
    // key：json数组的key
    //返回类型是json数组的字符串
    const char* vectorToString(string key, vector<string> vec)
    {
        Document doc;
        doc.SetObject();

        Document::AllocatorType& allocator = doc.GetAllocator();
        Value array(kArrayType);

        for(vector<string>::iterator iter = vec.begin();iter != vec.end();iter++){
            string str(*iter);
            Value v(str.c_str(),allocator);
            array.PushBack(v,allocator);
        }
        Value valName(key.c_str(),allocator);
        doc.AddMember(valName,array,allocator);

        return documentToString(&doc);
    }

    int doGetUserStatusByUserAddr(string strUserManagerAddr, string userAddress)
    {
        bcwasm::Address umAddr(strUserManagerAddr);
        bcwasm::DeployedContract userManagerContract(umAddr);
        string  strAccount = userManagerContract.callString("getAccountByAddress",userAddress);
        if ("" == strAccount){
            return 2; //找不到平台用户信息
        }

        Document doc;
        if ( doc.Parse(strAccount.c_str()).HasParseError() )
        {
            return 3; //解析json字符串失败
        }

        unsigned int code = doc["code"].GetUint();
        if ( 0  != code)
        {
            bcwasm::println(doc["msg"].GetString());
            return 4; //调用接口失败
        }

        const Value &rjObject = doc["data"];
        if (!rjObject.IsObject())
        {
            return 5;
        }

        if (!rjObject.HasMember("status"))
        {
            return 6;
        }

        if(!rjObject["status"].IsUint())
        {
            return 7;
        }

        unsigned int status = rjObject["status"].GetUint();
        if (1 == status){
            return 8;
        }else if( 2 == status){
            return 9;
        }

        return  0;
    }

    int doGetUserStatus(string strUserManagerAddr)
    {
        string userAddress = "0x" + bcwasm::origin().toString();
        return doGetUserStatusByUserAddr(strUserManagerAddr, userAddress);
    }

    //根据调用者地址获取用户状态
    // 返回值含义：
    // 0： 成功
    // 非零： 失败
    int getUserStatus()
    {
        bcwasm::DeployedContract regManagerContract(regManagerAddr);
        string  strUserManagerAddr = regManagerContract.callString("getContractAddress","__sys_UserManager","latest");
        if ("" == strUserManagerAddr){
            return  1; //找不到用户管理合约的地址
        }

        return  doGetUserStatus(strUserManagerAddr);
    }

    //根据用户地址获取用户状态
    // 返回值含义：
    // 0： 成功
    // 非零： 失败
    int getUserStatusByUserAddr(string userAddress)
    {
        bcwasm::DeployedContract regManagerContract(regManagerAddr);
        string  strUserManagerAddr = regManagerContract.callString("getContractAddress","__sys_UserManager","latest");
        if ("" == strUserManagerAddr){
            return  1; //找不到用户管理合约的地址
        }

        return  doGetUserStatusByUserAddr(strUserManagerAddr, userAddress);
    }

    void doGetRoles(string strRoleManagerAddr, vector<string> &roles)
    {
        string userAddress = "0x" + bcwasm::origin().toString();
        bcwasm::Address rmAddr(strRoleManagerAddr);
        bcwasm::DeployedContract roleManagerContract(rmAddr);
        string strRoles = roleManagerContract.callString("getRolesByAddress", userAddress.c_str());
        if ("" == strRoles) {
            return; //空vector
        }

        Document doc;
        doc.Parse(strRoles.c_str());
        if(doc.HasMember("code") && 0 != doc["code"].GetInt()) {
            bcwasm::println("[util] [doGetRoles] no valid roles.");
            return;
        }
        if(!doc["data"].IsArray()) {
            bcwasm::println("[util] [doGetRoles] invaid roles format.");
            return;
        }
        for (Value::ConstValueIterator itr = doc["data"].Begin(); itr != doc["data"].End(); ++itr)
        {
            roles.push_back(itr->GetString());
        }

        return;
    }

    //根据用户地址获取用户角色
    void getRoles(vector<string> &roles)
    {
        bcwasm::DeployedContract regManagerContract(regManagerAddr);
        string strRoleManagerAddr = regManagerContract.callString("getContractAddress", "__sys_RoleManager","latest");
        if ("" == strRoleManagerAddr) {
            return; //空vector
        }

        return doGetRoles( strRoleManagerAddr, roles);
    }

    //check if the caller have permission to change requested roles
    //@callerRoles: the roles of caller
    //@roles: being modified roles
    bool ifHavePermission(vector<string> callerRoles , vector<string> roles)
        {
            for (auto const&  callerRole   :  callerRoles)
            {
                    if( CHAIN_CREATOR == callerRole )
                    {
                            bool havePermission(true);
                            for (auto const&  r   :  roles)
                            {
                                if( CHAIN_CREATOR == r )
                                {
                                    havePermission = false;
                                    break;
                                }
                            }

                        if(havePermission)
                        {
                            return true;
                        }

                    }else if( CHAIN_ADMIN == callerRole )
                    {
                        bool havePermission(true);
                        for (auto const&  r   :  roles)
                            {
                                if( CHAIN_CREATOR == r || CHAIN_ADMIN == r )
                                {
                                    havePermission = false;
                                    break;
                                }
                            }

                        if(havePermission)
                        {
                            return true;
                        }

                    }else if(CONTRACT_ADMIN == callerRole  || NODE_ADMIN == callerRole  || CONTRACT_DEPLOYER == callerRole)
                    {
                            bool havePermission(true);
                            for (auto const&  r   :  roles)
                            {
                                if( CHAIN_CREATOR == r || CHAIN_ADMIN == r || NODE_ADMIN == r || CONTRACT_ADMIN == r || CONTRACT_DEPLOYER == r)
                                {
                                        havePermission = false;
                                        break;
                                }
                            }

                        if(havePermission)
                        {
                            return true;
                        }
                    }
            }

            return false;               
        }

        //check if the role is valid
        bool isRoleValid(string  role)
        {
            if ("" == role)
            {
                return false;
            }
        
            if ( role != CHAIN_CREATOR &&   role != CHAIN_ADMIN &&
                    role != NODE_ADMIN &&    role != CONTRACT_ADMIN &&
                    role != CONTRACT_DEPLOYER) 
            {
                return false;
            }
    
            return true;
        }

        //check if the roles is valid
        bool isRolesValid(vector<string> roles)
        {
            if (roles.empty())
            {
                return false;
            }

            for (auto const&  r : roles)
            {
                if (r != util::CHAIN_CREATOR &&  r != util::CHAIN_ADMIN &&
                    r != util::NODE_ADMIN && r != util::CONTRACT_ADMIN &&
                    r != util::CONTRACT_DEPLOYER )
                {
                    return false;
                }
            }

            return true;
        }

        string toLower(const string& str)
        {
            string tmp = str;
            transform(tmp.begin(), tmp.end(), tmp.begin(), ::tolower);
            return tmp;
        }

        string toUpper(const string& str)
        {
            string tmp = str;
            transform(tmp.begin(), tmp.end(), tmp.begin(), ::toupper);
            return tmp;
        }

        const char* makeReturnedStr(const string& ret)
        {
            char* ptr = new char[ret.size() + 1];
            strcpy(ptr, ret.c_str());
            return ptr;
        }

        bool isValidUser(const string& addr)
        {
            bcwasm::DeployedContract cnsManager(gdef::cnsManagerAddr);
            string userManagerAddr = cnsManager.callString("getContractAddress", gdef::userManager, "latest");

            //if initializing
            if (userManagerAddr.size() == 0)
            {
                return true;
            }

            bcwasm::DeployedContract b(userManagerAddr);
            if (b.callInt64("isValidUser", addr) == 1)
            {
                return true;
            }
            
            return false;
        }

        bool hasRole(const string& addr, const string& role)
        {
            bcwasm::DeployedContract cnsManager(gdef::cnsManagerAddr);
            string roleManagerAddr = cnsManager.callString("getContractAddress", gdef::roleManager, "latest");

            //if initializing
            if (roleManagerAddr.size() == 0)
            {
                return true;
            }

            bcwasm::DeployedContract b(roleManagerAddr);
            if (b.callInt64("hasRole", addr.c_str(), role.c_str()) == 1)
            {
                return true;
            }
            
            return false;
        }

        bool isAddress(const string& str)
        {
            string addr = toLower(str);

            //must start with 0x
            if (addr.size() != 42 || addr.substr(0, 2) != "0x")
                return false;

            for (int i = 2; i < 42; ++i)
            {
                if (!isxdigit(addr[i]))
                    return false;
            }

            return true;
        }

        bool isPublicKey(const string& str)
        {
            string key = toLower(str);

            if (key.substr(0, 2) == "0x")
                key = key.erase(0, 2);

            if (key.size() != 128)
                return false;

            for (int i = 0; i < 128; ++i)
            {
                if (!isxdigit(key[i]))
                    return false;
            }

            return true;
        }
}
