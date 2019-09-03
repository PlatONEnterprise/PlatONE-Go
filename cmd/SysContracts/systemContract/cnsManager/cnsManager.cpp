#include <stdlib.h>
#include <string.h>
#include <string>
#include <chrono>
#include <bcwasm/bcwasm.hpp>

#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

#include "../common/util.hpp"

#define RETURN_CHARARRAY(src, size)       \
    do                                    \
    {                                     \
        char *buf = (char *)malloc(size); \
        memset(buf, 0, size);             \
        strcpy(buf, src);                 \
        return buf;                       \
    } while (0)

namespace SystemContract
{

struct ContractInfo
{
    std::string name;    // 注册合约名
    std::string version; // 1.0.0.0
    std::string address; // 合约地址 0x...
    std::string origin;  // 创建者地址 0x... 暂保留，具体再讨论
    int64_t create_time;     // 合约创建时间
    bool enabled;        // CNS服务是否激活
    BCWASM_SERIALIZE(ContractInfo, (name)(version)(address)(origin)(create_time)(enabled));
};

char cnsMapName[] = "bcwasmCnsMap";
//TODO: 这里要用bcwasm::db::Map
typedef bcwasm::db::Map<cnsMapName, std::string, ContractInfo> cnsMap_t;
class CnsManager : public bcwasm::Contract
{
  public:
    CnsManager()
    {
        initErrCodes();
    }

    /// 实现父类: bcwasm::Contract 的虚函数
    /// 该函数在合约首次发布时执行，仅调用一次
    void init()
    {        
        bcwasm::DeployedContract regManagerContract("0x0000000000000000000000000000000000000011");
        std::string addr = bcwasm::address().toString();
        regManagerContract.call("updateCnsManager", addr, "1.0.0.1");
        BCWASM_EMIT_EVENT(init, "call updateCnsManager in init()", addr.c_str());
        bcwasm::println("CnsManager init success...");
    }

    BCWASM_EVENT(Notify, uint64_t, const char *)
    BCWASM_EVENT(init, const char *,  const char *)

  public:
    enum Code
    {
        SUCCESS=0,
        FAILURE
    };

  public:
    // 在其他合约的init()中注册合约
    // @name required
    // @version required
    // @return 含义如下：
    //  0 合约init（）内注册成功
    //  1 合约init（）内注册失败
    int cnsRegisterFromInit(const char *name, const char *version)
    {
        if (bcwasm::isFromInit() != 0) {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] cnsRegisterFromInit can only be called from init.");
            bcwasm::println("ERR : [CNS] cnsRegisterFromInit can only be called from init.");
            return FAILURE;
        }
        bcwasm::Address contractAddress = bcwasm::caller();
        std::string contractAddressStr = contractAddress.toString();
        const char *address = contractAddressStr.c_str();
        if (std::string(name).find("__sys_") == 0 && !canRegisterSysContract())
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Not allowed for registering system contract.");
            bcwasm::println("ERR : [CNS] Not allowed for registering system contract.");
            return FAILURE;
        }
        return doCnsRegister(name, version, address);
    }

    // 注册合约
    // @name required
    // @version required
    // @address required
    // @return 含义如下：
    //  0 合约外补注册成功
    //  1 合约外补注册失败
    int cnsRegister(const char *name, const char *version, const char *address)
    {
        // permission check
        if (bcwasm::isFromInit() == 0) {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] cnsRegister can't be called from init().");
            bcwasm::println("ERR : [CNS] cnsRegister can't be called from init().");
            return FAILURE;
        }

        std::string addr = address;
        util::formatAddress(addr);
        bcwasm::Address contractAddress(addr);

        if (bcwasm::isOwner(contractAddress, bcwasm::origin()) != 0)
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Not owner of registered contract.");
            bcwasm::println("ERR : [CNS] Not owner of registered contract.");
            return FAILURE;
        }
        if (std::string(name).find("__sys_") == 0 && !canRegisterSysContract())
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Not allowed for registering system contract.");
            bcwasm::println("ERR : [CNS] Not allowed for registering system contract.");
            return FAILURE;
        }

        return doCnsRegister(name, version, address);
    }

    // 注销特定合约
    // @name required
    // @version optional
    // @return 含义如下：
    //  0 注销成功
    //  1 注销失败
    int cnsUnregister(char *name, char *version)
    {
        // parameter check
        if (name == nullptr || version == nullptr || strlen(name) == 0 || strlen(version) == 0)
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "Input parameters are invalid.");
            bcwasm::println("ERR : [CNS] Input parameters are invalid.");
            return FAILURE;
        }

        std::string ver(version);
        std::transform(ver.begin(), ver.end(), ver.begin(), ::tolower);
        if (ver.compare("latest") == 0)
        {
            // get latest active version
            ver = getLatestVersion(name);
        }
        if (ver.empty())
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "Latest version is empty.");
            bcwasm::println("ERR : [CNS] Latest version is empty.");
            return FAILURE;
        }
        
        ContractInfo *cnsInfoPtr = cnsMap.find(std::string(name) + ver);
        if (nullptr != cnsInfoPtr)
        {

            // Only contract owner can unregister address
            bcwasm::Address originAddress = bcwasm::origin();
            bcwasm::Address contractAddress(cnsInfoPtr->address);
            if (bcwasm::isOwner(contractAddress, originAddress) != 0)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "Not owner of registered contract.");
                bcwasm::println("ERR : [CNS] Not owner of registered contract.");
                return FAILURE;
            }
            //cnsMap[std::string(name) + ver] = *cnsInfoPtr;
            cnsInfoPtr->enabled = false;
            cnsMap.update(std::string(name) + ver, *cnsInfoPtr);
            BCWASM_EMIT_EVENT(Notify, SUCCESS, "cnsUnregister succeed!");
            bcwasm::println("OK : [CNS] cnsUnregister succeed!");
            return SUCCESS;
        }
        else
        {
            // didn't register before, skip!
            BCWASM_EMIT_EVENT(Notify, FAILURE, "Name and version didn't register before.");
            bcwasm::println("ERR : [CNS] Name and version didn't register before.");
            return FAILURE;
        }
    }

    const char *getContractAddress(char *name, char *version) const
    {
        std::string addr;
        if (name == nullptr || version == nullptr || strlen(name) == 0 || strlen(version) == 0)
        {
            addr = "";
            RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
        }

        std::string ver(version);
        std::transform(ver.begin(), ver.end(), ver.begin(), ::tolower);
        if (ver.compare("latest") == 0)
        {
            ver = getLatestVersion(name);
        }
        if (ver.empty())
        {
            addr = "";
            RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
        }
        const ContractInfo *cnsInfoPtr = cnsMap.find(std::string(name) + ver);
        if (nullptr != cnsInfoPtr && cnsInfoPtr->enabled)
        {
            addr = cnsInfoPtr->address;
        }
        else
        {
            // the contract didn't register in CNS, so return 0x0 address
            //bcwasm::println("[getContractAddress] not found contract");
            addr = "";
        }
        RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
    }

    //获取所有已注册合约
    char *getRegisteredContracts(int pageNum, int pageSize) const
    {
        std::vector<ContractInfo> list;
        // parameter check
        if (pageNum < 0 || pageSize < 1)
        {
            return serializeContractsInfo(1, errCodes.at(1).c_str(), list);
        }

        int count = 0;
        int begin = pageNum * pageSize;
        int end = (pageNum + 1) * pageSize;
        //const std::set<string>& keys = cnsMap.getKeys();

        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            count++;
            if (count > end)
                break;
            if (count >= begin && count <= end)
            {
                ContractInfo info = it->second();
                if(info.enabled == true)
                {list.push_back(info);}
            }
        }

        return serializeContractsInfo(0, errCodes.at(0).c_str(), list);
    }

    //获取某人已注册合约
    char *getRegisteredContractsByAddress(char *origin, int pageNum, int pageSize) const
    {
        std::vector<ContractInfo> list;
        // parameter check
        std::string ori = origin;
        util::formatAddress(ori);

        if (origin == nullptr || strlen(origin) == 0 || !util::checkAddressFormat(ori.c_str()) || pageNum < 0 || pageSize < 1)
        {
            return serializeContractsInfo(1, errCodes.at(1).c_str(), list);
        }

        int count = 0;
        int begin = pageNum * pageSize;
        int end = (pageNum + 1) * pageSize;
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info = it->second();
            if (info.origin.compare(ori) == 0)
            {
                count++;
                if (count > end)
                break;
                if (count >= begin && count <= end)
                {
                    list.push_back(info);
                }
        }
            }
           

        return serializeContractsInfo(0, errCodes.at(0).c_str(), list);
    }

    // 是否已注册
    // @name required
    // @return 含义如下：
    //  0 未注册
    //  1 已注册
    int ifRegisteredByName(char *name) const
    {
        int activatedFlag = 0;
        if (name == nullptr || strlen(name) == 0 || !checkNameFormat(name))
        {
            return activatedFlag;
        }
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info = it->second();
            if (info.name.compare(name) == 0 && info.enabled)
            {
                
                activatedFlag = 1;
                break;
            }
        }
        return activatedFlag;
    }

    // 是否已注册
    // @address required
    // @return 含义如下：
    //  0 未注册
    //  1 已注册
    int ifRegisteredByAddress(char *address) const
    {
        
        int activatedFlag = 0;
        std::string addr = address;
        util::formatAddress(addr);

        if (address == nullptr || strlen(address) == 0 || !util::checkAddressFormat(addr.c_str()))
        {
            return activatedFlag;
        }
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info =it->second();
            if (info.address.compare(addr) == 0 && info.enabled)
            {
                activatedFlag = 1;
            }
        }
        return activatedFlag;
    }

    // 根据合约地址查询注册信息
    // @address required
    // @return 含义如下：
    //  0 未注册
    //  1 已注册
    char* getContractInfoByAddress(char *address) const
    {
        std::vector<ContractInfo> list;
        int activatedFlag = 0;
        std::string addr = address;
        util::formatAddress(addr);

        if (address == nullptr || strlen(address) == 0 || !util::checkAddressFormat(addr.c_str()))
        {
            return serializeContractsInfo(1, errCodes.at(1).c_str(), list);
        }
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info = it->second();
            if (info.address.compare(addr) == 0 && info.enabled)
            {
                activatedFlag = 1;
                list.push_back(info);
                break;
            }
        }
        if (activatedFlag){
            return serializeContractsInfo(0, errCodes.at(0).c_str(), list);
        }

        return serializeContractsInfo(1, errCodes.at(1).c_str(), list);        
    }

    // 查询历史合约(包含已注销合约)
    char *getHistoryContractsByName(char *name) const
    {
        std::vector<ContractInfo> list;
        if (name == nullptr || strlen(name) == 0 || !checkNameFormat(name))
        {
            return serializeContractsInfo(1, errCodes.at(1).c_str(), list);
        }
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info = it->second();
            if (info.name.compare(std::string(name)) == 0)
            {
                list.push_back(info);
            }
        }

        return serializeContractsInfo(0, errCodes.at(0).c_str(), list);
    }

  private:
    char *serializeContractsInfo(int code, const char *message, std::vector<ContractInfo> list) const
    {
        rapidjson::Document jsonDoc;                                            //生成一个dom元素Document
        rapidjson::Document::AllocatorType &allocator = jsonDoc.GetAllocator(); //获取分配器
        jsonDoc.SetObject();                                                    //将当前的Document设置为一个object，也就是说，整个Document是一个Object类型的dom元素

        int size = (int)sizeof(message);
        char *buf = (char *)malloc(size);
        memset(buf, 0, size);
        strcpy(buf, message);

        //添加属性
        jsonDoc.AddMember("code", code, allocator);

        rapidjson::Value msg(buf, allocator);
        jsonDoc.AddMember("msg", msg, allocator);

        rapidjson::Value *data = new rapidjson::Value(rapidjson::kObjectType);

        //生成一个object数组
        rapidjson::Value contractArray(rapidjson::kArrayType);

        for (std::vector<ContractInfo>::iterator it = list.begin(); it != list.end(); ++it)
        {
            rapidjson::Value contra(rapidjson::kObjectType);
            rapidjson::Value name(it->name.c_str(), allocator);
            rapidjson::Value version(it->version.c_str(), allocator);
            util::formatAddress(it->address);
            rapidjson::Value address(it->address.c_str(), allocator);
            util::formatAddress(it->origin);
            rapidjson::Value origin(it->origin.c_str(), allocator);

            contra.AddMember("name", name, allocator);
            contra.AddMember("version", version, allocator);
            contra.AddMember("address", address, allocator);
            contra.AddMember("origin", origin, allocator);
            contra.AddMember("create_time", it->create_time, allocator);
            contra.AddMember("enabled", it->enabled, allocator);
            contractArray.PushBack(contra, allocator); //添加到数组
        }
    
        int total = (int)list.size();
        data->AddMember("total", total, allocator);
        data->AddMember("contract", contractArray, allocator);

        jsonDoc.AddMember("data", *data, allocator);

        //生成字符串
        rapidjson::StringBuffer buffer;
        rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
        jsonDoc.Accept(writer);

        std::string strJson = buffer.GetString();
        RETURN_CHARARRAY(strJson.c_str(), strJson.size() + 1);
    }

    // parameters check functions
  private:
    bool checkNameFormat(const char *n) const
    {
        if (n == nullptr)
            return false;
        int length = strlen(n);
        if (length < util::CONTRACT_NAME_LENGTH_MIN || length > util::CONTRACT_NAME_LENGTH_MAX)
            return false;

        // check first character
        if ((*n >= 'a' && *n <= 'z') || (*n >= 'A' && *n <= 'Z') || *n == '_')
            return true;

        return false;
    }

    std::vector<std::string> spiltVersion(const char *v) const
    {
        if (v == nullptr)
            return std::vector<std::string>();
        if (strlen(v) == 0)
            return std::vector<std::string>();

        std::string exp = v;
        char delimiter = '.';
        std::vector<std::string> arr;
        std::string acc = "";
        for (int i = 0; i < exp.size(); i++)
        {
            if (exp[i] == delimiter)
            {
                arr.push_back(acc);
                acc = "";
            }
            else
                acc += exp[i];
        }

        if (acc.length() != 0)
            arr.push_back(acc);

        return arr;
    }

    bool checkVersionFormat(const char *v) const
    {
        if (v == nullptr)
            return false;
        if (strlen(v) == 0)
            return false;

        std::vector<std::string> ret = spiltVersion(v);
        // standard format: 0.0.0.0
        if (ret.size() != 4)
            return false;

        for (int i = 0; i < ret.size(); i++)
        {
            for (int j = 0; j < ret[i].length(); j++)
            {
                if (!isdigit(ret[i][j]))
                    return false;
            }
        }
        return true;
    }

    int str2Int(const char *str) const
    {
        int res = 0;
        for (int i = 0; str[i] != '\0'; ++i)
            res = res * 10 + str[i] - '0';

        return res;
    }

    int compareVersion(const char *v1, const char *v2) const
    {
        // version format : X.X.X.X
        // please make sure v1,v2 format is valid before call this function
        std::vector<std::string> ret1 = spiltVersion(v1);
        std::vector<std::string> ret2 = spiltVersion(v2);

        for (int i = 0; i < 4; i++)
        {
            if (str2Int(ret1[i].c_str()) > str2Int(ret2[i].c_str()))
            {
                return 1;
            }
            else if (str2Int(ret1[i].c_str()) < str2Int(ret2[i].c_str()))
            {
                return -1;
            }
        }
        return 0;
    }
    char *getLatestVersion(const char *name) const
    {
        std::string latest = "0.0.0.0";
        if (name == nullptr || strlen(name) == 0 || !checkNameFormat(name)) {
            RETURN_CHARARRAY(latest.c_str(), latest.size() + 1);
        }
        for(auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
        {
            ContractInfo info = it->second();
            if (info.name.compare(std::string(name)) == 0 && compareVersion(info.version.c_str(), latest.c_str()) == 1)
            {
                latest = info.version;
            }
        }

        RETURN_CHARARRAY(latest.c_str(), latest.size() + 1);
    }

    //注册申请者是否可以注册系统合约
    bool canRegisterSysContract()
    {
        // get user status
        std::string strUserManagerAddr = getContractAddress("__sys_UserManager", "latest");
        if (strUserManagerAddr.empty())
        {
            //如果找不到用户管理合约的地址，则直接pass
            return true;
        }

        if (util::doGetUserStatus(strUserManagerAddr) != 0)
            return false; //获取用户状态失败

        // get user roles
        std::string strRoleManagerAddr = getContractAddress("__sys_RoleManager", "latest");
        if (strRoleManagerAddr.empty())
        {
            //如果找不到用户角色管理合约的地址，则直接pass
            return true;
        }
        std::vector<std::string> roles;
        util::doGetRoles(strRoleManagerAddr, roles);

        for (vector<string>::iterator iter = roles.begin(); iter != roles.end(); iter++)
        {
            string str(*iter);
            if (str == "chainCreator")
                return true;
        }
        return false;
    }

    void initErrCodes()
    {
        // status 0: ok
        errCodes[0] = "ok";

        // status 1: input is invalid
        errCodes[1] = "input is invalid";

        // status 2: internal error
        errCodes[2] = "internal error";

        // status 3: not registered
        errCodes[3] = "not registered";

        // status 4: already unregistered
        errCodes[4] = "already unregistered";
    }

    int doCnsRegister(const char *name, const char *version, const char *address)
    {
        // parameters check
        // bcwasm::getState(TOTAL,total);
        if (!checkNameFormat(name))
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Name format is invalid.");
            bcwasm::println("ERR : [CNS] Name format is invalid.");
            return FAILURE;
        }

        if (!checkVersionFormat(version))
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Version format is invalid.");
            bcwasm::println("ERR : [CNS] Version format is invalid.");
            return FAILURE;
        }

        std::string addr = address;
        util::formatAddress(addr);

        if (!util::checkAddressFormat(addr.c_str()))
        {
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Address format is invalid.");
            bcwasm::println("ERR : [CNS] Address format is invalid.");
            return FAILURE;
        }

        std::string ori = bcwasm::origin().toString();
        util::formatAddress(ori);

        ContractInfo info = {0};
        info.name = std::string(name);
        info.version = std::string(version);
        info.address = addr;
        info.origin = ori;
        info.create_time = timestamp();
        info.enabled = true;
        ContractInfo *cnsInfoPtr;
        cnsInfoPtr = cnsMap.find(std::string(name) + version);
        if (nullptr != cnsInfoPtr)
        {
            // this (name + version) already registered and activated in CNS, skip!
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Name and version is already registered and activated in CNS.");
            bcwasm::println("ERR : [CNS] Name and version is already registered and activated in CNS.");
            return FAILURE;
        }
        else
        { 
            for (auto it = cnsMap.begin(); it != cnsMap.end(); ++it)
            {
                ContractInfo tmp = it->second();
                if (tmp.name == std::string(name) && tmp.origin.compare(ori) != 0)
                {
                    BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Name is already registered");
                    bcwasm::println("ERR : [CNS] Name is already registered");
                    return FAILURE;
                }
            }

            char *curLatestVersion = getLatestVersion(name);
            if (compareVersion(version, curLatestVersion) == 1)
            {
                // Monotonically increasing version
                cnsMap.insert(std::string(name) + std::string(version),info);
                BCWASM_EMIT_EVENT(Notify, SUCCESS, "cnsRegister succeed!");
                return SUCCESS;
            }
            else
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CNS] Version must be larger than current latest versoin.");
                bcwasm::println("ERR : [CNS] Version must be larger than current latest versoin.");
                return FAILURE;
            }
        }
    }

  private:
    cnsMap_t cnsMap;
    std::map<int, std::string> errCodes;
};
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用

BCWASM_ABI(SystemContract::CnsManager, cnsRegisterFromInit)
BCWASM_ABI(SystemContract::CnsManager, cnsRegister)
BCWASM_ABI(SystemContract::CnsManager, cnsUnregister)
BCWASM_ABI(SystemContract::CnsManager, getContractAddress)
BCWASM_ABI(SystemContract::CnsManager, getRegisteredContracts)
BCWASM_ABI(SystemContract::CnsManager, getRegisteredContractsByAddress)
BCWASM_ABI(SystemContract::CnsManager, ifRegisteredByName)
BCWASM_ABI(SystemContract::CnsManager, ifRegisteredByAddress)
BCWASM_ABI(SystemContract::CnsManager, getHistoryContractsByName)
BCWASM_ABI(SystemContract::CnsManager, getContractInfoByAddress)
//bcwasm autogen begin
extern "C" { 
int cnsRegisterFromInit(const char * name,const char * version) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.cnsRegisterFromInit(name,version);
}
int cnsRegister(const char * name,const char * version,const char * address) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.cnsRegister(name,version,address);
}
int cnsUnregister(char * name,char * version) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.cnsUnregister(name,version);
}
const char * getContractAddress(char * name,char * version) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.getContractAddress(name,version);
}
char * getRegisteredContracts(int pageNum,int pageSize) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.getRegisteredContracts(pageNum,pageSize);
}
char * getRegisteredContractsByAddress(char * origin,int pageNum,int pageSize) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.getRegisteredContractsByAddress(origin,pageNum,pageSize);
}
int ifRegisteredByName(char * name) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.ifRegisteredByName(name);
}
int ifRegisteredByAddress(char * address) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.ifRegisteredByAddress(address);
}
char * getContractInfoByAddress(char * address) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.getContractInfoByAddress(address);
}
char * getHistoryContractsByName(char * name) {
SystemContract::CnsManager CnsManager_bcwasm;
return CnsManager_bcwasm.getHistoryContractsByName(name);
}
void init() {
SystemContract::CnsManager CnsManager_bcwasm;
CnsManager_bcwasm.init();
}

}
//bcwasm autogen end