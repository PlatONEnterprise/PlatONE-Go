//auto create contract
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
    std::string version; // 1.0.0.0
    std::string address; // 合约地址 0x...
    std::string origin;  // 创建者地址 0x... 暂保留，具体再讨论
    BCWASM_SERIALIZE(ContractInfo, (version)(address)(origin));
};

char cnsMapName[] = "bcwasmCnsMap";
char historyMapName[] = "historyMap";
char currentManagerKey[] = "currentManagerName";
char currentVersionKey[] = "currentVersion";

//TODO: 这里要用bcwasm::db::Map
typedef bcwasm::db::Map<cnsMapName, std::string, ContractInfo> cnsMap_t;

class CnsProxy : public bcwasm::Contract
{
  public:
    CnsProxy()
    {
        initErrCodes();
    }

    /// Event
    BCWASM_EVENT(Notify, uint64_t, const char *)
    BCWASM_EVENT(Init, uint64_t, const char *)
    BCWASM_EVENT(UpdateCnsManager, uint64_t, const char * , const char* , const char* )
    BCWASM_EVENT(getCnsManagerAddress, uint64_t, const char * , const char*)

    /// 实现父类: bcwasm::Contract 的虚函数
    /// 该函数在合约首次发布时执行，仅调用一次
    void init()
    {
        BCWASM_EMIT_EVENT(Init,0, "[CnsProxy]: Init CnsProxy ...");
    }

    public:
    enum Code
    {
        SUCCESS=0,
        FAILURE
    };

    // address: "0x123..."
    // version: "0.0.0.1"
    int updateCnsManager(const char *address, const char *version)
    {
        std::string addr = address;
        util::formatAddress(addr);
        bcwasm::Address contractAddress(addr);
        if (bcwasm::isFromInit() != 0) {
            if (bcwasm::isOwner(contractAddress, bcwasm::origin()) != 0)
            {
                BCWASM_EMIT_EVENT(UpdateCnsManager, FAILURE, "ERR : [CnsProxy] Not owner of registered contract.", address, version);
                return FAILURE;
            }
        }

        if(checkCallerPermission() == false){
            BCWASM_EMIT_EVENT(UpdateCnsManager, FAILURE, "ERR : [CnsProxy] checkCallerPermission failed.", address, version);
            return FAILURE;
        }

        int ret = doCnsManagerRegister(address, version);
        if (ret != 0){
            BCWASM_EMIT_EVENT(UpdateCnsManager, FAILURE, "ERR : [CnsProxy] doCnsManagerRegister failed.", address, version);
            return FAILURE;
        }

        bcwasm::setState(currentManagerKey, addr);
        bcwasm::setState(currentVersionKey, version);

        BCWASM_EMIT_EVENT(UpdateCnsManager, SUCCESS, "[CnsProxy]: UpdateCnsManager Success", address, version);
        return SUCCESS;
    }

    const char *getCnsManagerAddress(char *version) const
    {
        std::string addr;
        if (version == nullptr || strlen(version) == 0)
        {
            addr = "";
            RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
        }

        std::string ver(version);
        std::transform(ver.begin(), ver.end(), ver.begin(), ::tolower);

        if (ver.compare("latest") == 0)
        {
            bcwasm::getState(currentManagerKey, addr);
            RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
        }

        if (ver.empty())
        {
            addr = "";
            RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
        }

        const ContractInfo *cnsInfoPtr = cnsMap.find(ver);

        if (nullptr != cnsInfoPtr)
        {
            addr = cnsInfoPtr->address;
        }
        else
        {
            addr = "";
        }
        RETURN_CHARARRAY(addr.c_str(), addr.size() + 1);
    }

    // 获取所有已注册合约
    char *getAllRegisteredCnsManager(int pageNum, int pageSize) const
    {
        std::vector<ContractInfo> list;
        // parameter check
        if (pageNum < 0)
        {
            return serializeContractsInfo(1, errCodes.at(1).c_str(), list);
        }

        // minimal pageSize is 5
        if (pageSize < 5){
            pageSize = 5;
        }

        int count = 0;
        int begin = pageNum * pageSize;
        int end = (pageNum + 1) * pageSize;
        const std::set<string>& keys = cnsMap.getKeys();
        for(auto it = keys.begin(); it != keys.end(); ++it)
        {
            count++;
            if (count > end)
                break;
            if (count >= begin && count <= end)
            {
                ContractInfo info = *(cnsMap.find(*it));
                list.push_back(info);
            }
        }
        return serializeContractsInfo(0, errCodes.at(0).c_str(), list);
    }
 
  private:
    bool checkCallerPermission(){
        std::string cnsManageraddr;
        bcwasm::getState(currentManagerKey, cnsManageraddr);
        if(cnsManageraddr.empty()) {
            // 若cnsManager还未注册，则默认有权限修改
            // bcwasm::println("cnsManager未注册");
            return true;
        }

        // get user status
        bcwasm::DeployedContract a("0x0000000000000000000000000000000000000011");
		std::string strUserManagerAddr = a.callString("getContractAddress", "__sys_UserManager", "latest");
        if (strUserManagerAddr.empty())
        {
            //如果找不到用户管理合约的地址，则直接pass
            // bcwasm::println("找不到用户管理合约地址");
            return true;
        }

        if (util::doGetUserStatus(strUserManagerAddr) != 0){
            bcwasm::println("用户状态出错");
            return false; //获取用户状态失败
        }

        // get user roles
        std::string strRoleManagerAddr = a.callString("getContractAddress", "__sys_RoleManager", "latest");
        if (strRoleManagerAddr.empty())
        {
            //如果找不到用户角色管理合约的地址，则直接pass
            // bcwasm::println("找不到角色管理合约地址");
            return true;
        }

        std::vector<std::string> roles;
        util::doGetRoles(strRoleManagerAddr, roles);

        for (vector<string>::iterator iter = roles.begin(); iter != roles.end(); iter++)
        {
            string str(*iter);
            if (str == util::CHAIN_CREATOR || str ==util::CHAIN_ADMIN){
                return true;
            }
                
        }
        return false;
    }
    
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
            // rapidjson::Value name(it->name.c_str(), allocator);
            rapidjson::Value version(it->version.c_str(), allocator);
            util::formatAddress(it->address);
            rapidjson::Value address(it->address.c_str(), allocator);
            util::formatAddress(it->origin);
            rapidjson::Value origin(it->origin.c_str(), allocator);

            // contra.AddMember("name", name, allocator);
            contra.AddMember("version", version, allocator);
            contra.AddMember("address", address, allocator);
            contra.AddMember("origin", origin, allocator);
            // contra.AddMember("create_time", it->create_time, allocator);
            // contra.AddMember("enabled", it->enabled, allocator);
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
        // if (name == nullptr || strlen(name) == 0 || !checkNameFormat(name)) {
        //     RETURN_CHARARRAY(latest.c_str(), latest.size() + 1);
        // }
        const std::set<string>& keys = cnsMap.getKeys();
        for(auto it = keys.begin(); it != keys.end(); ++it)
        {
            ContractInfo info = *(cnsMap.find(*it));
            if (compareVersion(info.version.c_str(), latest.c_str()) == 1)
            {
                latest = info.version;
            }
        }

        RETURN_CHARARRAY(latest.c_str(), latest.size() + 1);
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

    int doCnsManagerRegister(const char *address, const char *version)
    {
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
        info.version = std::string(version);
        info.address = addr;
        info.origin = ori;
        // info.create_time = timestamp();
        // info.enabled = true;

        std::map<std::string, ContractInfo>::iterator it;
        ContractInfo *cnsInfoPtr = cnsMap.find(version);
        if (nullptr != cnsInfoPtr)
        {
            // this (name + version) already registered and activated in CNS, skip!
            BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CnsProxy] Version is already registered in CnsProxy.");
            return FAILURE;
        }
        else
        { 
            char *curLatestVersion = getLatestVersion(version);
            if (compareVersion(version, curLatestVersion) == 1)
            {
                // Monotonically increasing version
                // cnsMap[std::string(name) + std::string(version)] = info;
                cnsMap.insert(std::string(version),info);
                BCWASM_EMIT_EVENT(Notify, SUCCESS, "cnsRegister succeed!");
                return SUCCESS;
            }
            else
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "ERR : [CnsProxy] Version must be larger than current latest versoin.");
                bcwasm::println("ERR : [CNS] Version must be larger than current latest versoin.");
                return FAILURE;
            }
        }
    }

  private:
    cnsMap_t cnsMap;
    cnsMap_t historyMap;
    std::map<int, std::string> errCodes;
};
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::CnsProxy, updateCnsManager)
BCWASM_ABI(SystemContract::CnsProxy, getCnsManagerAddress)
BCWASM_ABI(SystemContract::CnsProxy, getAllRegisteredCnsManager)

//bcwasm autogen begin
extern "C" { 
int updateCnsManager(const char * address,const char * version) {
SystemContract::CnsProxy CnsProxy_bcwasm;
return CnsProxy_bcwasm.updateCnsManager(address,version);
}
const char * getCnsManagerAddress(char * version) {
SystemContract::CnsProxy CnsProxy_bcwasm;
return CnsProxy_bcwasm.getCnsManagerAddress(version);
}
char * getAllRegisteredCnsManager(int pageNum,int pageSize) {
SystemContract::CnsProxy CnsProxy_bcwasm;
return CnsProxy_bcwasm.getAllRegisteredCnsManager(pageNum,pageSize);
}
void init() {
SystemContract::CnsProxy CnsProxy_bcwasm;
CnsProxy_bcwasm.init();
}

}
//bcwasm autogen end