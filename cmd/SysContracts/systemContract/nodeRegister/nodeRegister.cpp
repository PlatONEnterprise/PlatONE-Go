#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

#include <iostream>
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
char mapName[] = "RegisterInfo";

struct RegisterInfo
{
    std::string name;      //节点名称
    std::string owner;     //申请者的地址
    std::string desc;      //节点描述,长度限制1000
    int type;              //1:共识节点；0:观察者节点
    std::string publicKey; //公钥
    std::string externalIP;
    std::string internalIP;
    int rpcPort;
    int p2pPort;
    int status;           //0:未处理；1:申请通过；2:拒绝申请
    bool root;            //true 根节点 false 非跟节点
    std::string approver; //审核人的地址
    int64_t registerTime; //申请时间
    BCWASM_SERIALIZE(RegisterInfo, (name)(owner)(desc)(type)(publicKey)(externalIP)(internalIP)(rpcPort)(p2pPort)(status)(root)(approver)(registerTime));
};
//以publicKey为key，RegisterInfo整体为value
typedef bcwasm::db::Map<mapName, std::string, RegisterInfo> nodeMap_t;

class NodeRegister : public bcwasm::Contract
{
  private:
    nodeMap_t registerInfos;

  public:
    enum Code
    {
        SUCCESS = 0,
        BAD_PARAMETER,
        NO_PERMISSION
    };
    enum NodeType
    {
        OBSERVER = 0,
        CONSENSUS_NODE
    };
    enum GetRegisterInfoByStatus
    {
        UN_APPROVE = 0,
        PASSED,
        REFUSED
    };

    NodeRegister() {}

    void init()
    {
        //注册cns服务
        util::registerContractFromInit("__sys_NodeRegister", "1.0.0.0");
        bcwasm::println("NodeRegister init success...");
    }

    // 定义Event
    BCWASM_EVENT(Notify, uint64_t, const char *)

  public:
    // 申请节点
    // @nodeJson required
    // @return 含义如下：
    //  0 申请成功
    //  1 参数错误
    //  2 没有权限
    int registerNode(const char *nodeJson)
    {
        RegisterInfo info;

        if (nodeJson == nullptr || strlen(nodeJson) == 0)
        {
            bcwasm::println("nodeJson invalid");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: nodeJson invalid");
            return BAD_PARAMETER;
        }

        if (!deserializeNodeJson(nodeJson, info))
        {
            bcwasm::println("Parse node info error");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: Parse node info error");
            return BAD_PARAMETER;
        }

        if (info.desc.length() < 4 || info.desc.length() > 1000)
        {
            bcwasm::println("Description length illegal, more than 3 and less than 1000");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: description length illegal, more than 3 and less than 1000");
            return BAD_PARAMETER;
        }

        if (info.publicKey.length() < 4 || info.publicKey.length() > 1000)
        {
            bcwasm::println("publicKey length illegal, more than 3 and less than 1000");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: publicKey length illegal, more than 3 and less than 1000");
            return BAD_PARAMETER;
        }

        if ((info.type != OBSERVER) && (info.type != CONSENSUS_NODE))
        {
            bcwasm::println("type invalid");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: type invalid");
            return BAD_PARAMETER;
        }

        info.owner = getOrigin(); //owner
        info.registerTime = timestamp();

        //判断call的角色权限，只有链管理员和节点管理员可以申请节点
        if (!canRegister())
        {
        std:
            string message = info.owner + "have not permission to register node!";
            bcwasm::println(message);
            BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, message.c_str());
            return NO_PERMISSION;
        }

        //Publickey不可以重复
        if (ifPublickeyRegistered(info.publicKey))
        {
            bcwasm::println("Publickey already registered");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: Publickey already registered");
            return BAD_PARAMETER;
        }

        //checkNameFormat
        if (!checkNameFormat(info.name.c_str()))
        {
            bcwasm::println("name invalid，can only contain numbers, characters, _, -");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "register: name invalid，can only contain numbers, characters, _, -");
            return BAD_PARAMETER;
        }

        //节点名称不可以重复
        if (ifNameUsed(info.name))
        {
            bcwasm::println("Name already used");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "Name already used");
            return BAD_PARAMETER;
        }

        if (ifIPPortUsed(info.externalIP, info.internalIP, info.rpcPort, info.p2pPort))
        {
            bcwasm::println("IP:Port already Used!");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "IP:Port already Used!");
            return BAD_PARAMETER;
        }

        registerInfos.insert(info.publicKey, info);

        std::string msg = "Register node. Publickey: " + info.publicKey + ", name: " + info.name;
        bcwasm::println(msg);
        BCWASM_EMIT_EVENT(Notify, SUCCESS, msg.c_str());
        return SUCCESS;
    }

    // publicKey repuired
    // status repaired 1：审核通过，2：拒绝
    // @return 含义如下：
    //  0 审核成功
    //  1 参数错误
    //  2 没有权限
    int approve(char *publicKey, int status)
    {
        int result;
        RegisterInfo info;
        std::string nodeJson;

        if (publicKey == nullptr || strlen(publicKey) < 4 || strlen(publicKey) > 1000)
        {
            bcwasm::println("publicKey invalid");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "approve: publicKey invalid");
            return BAD_PARAMETER;
        }

        //判断approve权限，只有链管理员可以申请节点
        if (!canApprove())
        {
            std::string message = info.owner + "have not permission to register node!";
            bcwasm::println(message);
            BCWASM_EMIT_EVENT(Notify, NO_PERMISSION, message.c_str());
            return NO_PERMISSION;
        }

        if (status != PASSED && status != REFUSED)
        {
            bcwasm::println("approve:status parameter invalid");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "approve: status parameter invalid");
            return BAD_PARAMETER;
        }

        std::string pk = std::string(publicKey);
        std::transform(pk.begin(), pk.end(), pk.begin(), ::tolower);

        const RegisterInfo *InfoPtr = registerInfos.find(pk);
        if(nullptr != InfoPtr) 
        {
            info = *(registerInfos.find(pk));
            if (info.status != UN_APPROVE)
            {
                bcwasm::println("approve: status already set.");
                BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "approve: status already set");
                return BAD_PARAMETER;
            }
            if (ifNameUsed(info.name))
            {
                bcwasm::println("approve: Name already Used!");
                BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "approve: Name already Used!");
                return BAD_PARAMETER; //节点名称不可以重复
            }
            if (ifIPPortUsed(info.externalIP, info.internalIP, info.rpcPort, info.p2pPort))
            {
                bcwasm::println("approve IP:Port already Used!");
                BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "approve IP:Port already Used!");
                return BAD_PARAMETER; //IP + Port不可以重复
            }
            else
            {
                //更改状态信息并设置审核人
                info.status = status;
                info.approver = getOrigin();

                nodeJson = serializeNodeInfo(info);
                bcwasm::println("nodeJson: " + nodeJson);

                //add node info to nodeManager when status is Passed
                if (status == PASSED)
                {
                    bcwasm::println("begin call cnsManager");
                    //通过合约管理合约获取 nodeManager 合约地址
                    bcwasm::DeployedContract cnsManagerContract(gdef::cnsManagerAddr);
                    string strnodeManagerAddr = cnsManagerContract.callString("getContractAddress", "__sys_NodeManager", "latest");
                    bcwasm::DeployedContract nodeManagerContract(strnodeManagerAddr);
                    bcwasm::println("nodeManager address: " + strnodeManagerAddr);

                    // //调用nodeManager的add方法，添加节点信息
                    int callResult = nodeManagerContract.callInt64("add", nodeJson.c_str());
                    if (callResult != 0)
                    {
                        bcwasm::println("nodeManagerContract ERROR");
                        BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "Call nodeManager Contract failed");
                        return BAD_PARAMETER;
                    }
                }

                //持久化数据
                registerInfos.insert(pk, info);
                std::string msg = "Approve node :" + nodeJson;
                BCWASM_EMIT_EVENT(Notify, SUCCESS, msg.c_str());
                return SUCCESS;
            }
        }
        else
        {
            bcwasm::println("Can't find publickey");
            BCWASM_EMIT_EVENT(Notify, BAD_PARAMETER, "Can't find publickey");
            return BAD_PARAMETER;
        }
    }

    // 查询接口
    const char *getRegisterInfoByStatus(int status, int pageNum, int pageSize) const
    {
        int count = 0;
        int begin = pageNum * pageSize + 1;
        int end = (pageNum + 1) * pageSize;
        std::vector<RegisterInfo> list;
        std::string msg = "SUCCESS";
        int code = 0;

        if (pageNum < 0 || pageSize < 1)
        {
            code = 2;
            msg = "Parameter error!";
            return serializeRegisterInfo(code, msg.c_str(), list);
        }

        for(auto it = registerInfos.begin(); it != registerInfos.end(); ++it)
        {

            if ((it->second()).status == status)
            {
                count++;
                if (count >= begin && count <= end)
                {
                    list.push_back(it->second());
                }
                else
                {
                    break;
                }
            }
        }

        if (list.size() == 0)
        {
            code = 1;
            msg = "Can't find info";
        }

        return serializeRegisterInfo(code, msg.c_str(), list);
    }

    const char *getRegisterInfoByPublicKey(char *publicKey) const
    {
        std::vector<RegisterInfo> list;
        RegisterInfo info;
        std::string msg = "SUCCESS";
        int code = 0;

        if (publicKey == nullptr || strlen(publicKey) == 0)
        {
            bcwasm::println("publicKey invalid");
            code = 1;
            msg = "publicKey invalid";
            return serializeRegisterInfo(code, msg.c_str(), list);
        }

        std::string pk = std::string(publicKey);
        std::transform(pk.begin(), pk.end(), pk.begin(), ::tolower);

        const RegisterInfo *InfoPtr = registerInfos.find(pk);
        if(nullptr != InfoPtr)
        {
            info = *(registerInfos.find(pk));
            list.push_back(info);
        }

        if (list.size() == 0)
        {
            code = 1;
            msg = "Can't find info";
        }

        return serializeRegisterInfo(code, msg.c_str(), list);
    }

    const char *getRegisterInfoByOwnerAddress(char *owner) const
    {
        std::vector<RegisterInfo> list;
        std::string msg = "SUCCESS";
        int code = 0;

        if (owner == nullptr || strlen(owner) == 0)
        {
            bcwasm::println("owner invalid");
            code = 1;
            msg = "owner invalid";
            return serializeRegisterInfo(code, msg.c_str(), list);
        }

        std::string own = owner;
        util::formatAddress(own);

        for(auto it = registerInfos.begin(); it != registerInfos.end(); ++it)
        {
            if ((it->second()).owner.compare(own) == 0)
                list.push_back(it->second());
        }

        if (list.size() == 0)
        {
            code = 1;
            msg = "Can't find info";
        }

        return serializeRegisterInfo(code, msg.c_str(), list);
    }

  private:
    //获取合约调用者
    std::string getOrigin()
    {
        std::string origin = bcwasm::origin().toString();
        util::formatAddress(origin);
        return origin;
    }

    //调用者是否有申请权限
    bool canRegister()
    {
        if (util::getUserStatus() != 0)
            return false;

        vector<string> roles;
        util::getRoles(roles);

        for (vector<string>::iterator iter = roles.begin(); iter != roles.end(); iter++)
        {
            string str(*iter);
            if (str == "chainCreator" || str == "chainAdmin" || str == "nodeAdmin")
                return true;
        }
        return false;
    }

    //调用者是否有审核权限
    bool canApprove()
    {
        if (util::getUserStatus() != 0)
            return false;

        std::vector<string> roles;
        util::getRoles(roles);

        for (vector<string>::iterator iter = roles.begin(); iter != roles.end(); iter++)
        {
            string str(*iter);
            if (str == "chainCreator" || str == "chainAdmin")
                return true;
        }
        return false;
    }

    // Publickey是否已注册，以前的注册信息中使用过的Publickey不能再次申请
    // @return 含义如下：
    //  0 未申请
    //  1 已申请
    bool ifPublickeyRegistered(std::string publicKey)
    {
        const RegisterInfo *InfoPtr = registerInfos.find(publicKey);
        if(nullptr != InfoPtr)
        {
            return true;
        }
        else
        {
            return false;
        }
    }

    // Name是否已注册并审核通过，没有注册过的和未审核通过的可以再次申请
    // @return 含义如下：
    //  0 未申请
    //  1 已申请
    bool ifNameUsed(std::string name)
    {
        for(auto it = registerInfos.begin(); it != registerInfos.end(); ++it)
        {
            if ((it->second()).status == 1)
            {
                if ((it->second()).name == name)
                {
                    return true;
                }
            }
        }
        return false;
    }

    // IP和端口是否已注册并审核通过，没有注册过的和未审核通过的可以再次申请
    // @return 含义如下：
    //  0 未申请
    //  1 已申请
    bool ifIPPortUsed(std::string eIP, std::string iIP, int rport, int pport)
    {
        for(auto it = registerInfos.begin(); it != registerInfos.end(); ++it) 
        {
            if ((it->second()).status == 1)
            {
                int rpcPort = (it->second()).rpcPort;
                int p2pPort = (it->second()).p2pPort;
                std::string externalIP = (it->second()).externalIP;
                std::string internalIP = (it->second()).internalIP;

                if (externalIP == eIP || externalIP == iIP || internalIP == eIP || internalIP == iIP)
                {
                    if ((rpcPort == rport) || (rpcPort == pport) || (p2pPort == rport) || (p2pPort == pport))
                        return true;
                }
            }
        }
        return false;
    }

    bool checkNameFormat(const char *name) const
    {
        if (name == nullptr || strlen(name) < 0 || strlen(name) > 1000)
            return false;

        for (int i = 0; i < strlen(name); i++)
        {
            char ch = name[i];
            if ((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '-') || (ch == '_'))
                continue;
            else
                return false;
        }
        return true;
    }

    //Parse node info
    bool deserializeNodeJson(const std::string &nodejson, RegisterInfo &info)
    {
        Document doc;
        doc.Parse(nodejson.c_str());

        if (doc.HasParseError())
        {
            ParseErrorCode code = doc.GetParseError();
            DEBUG("Parse error, code：", code);
            return false;
        }

        Value::MemberIterator itr = doc.FindMember("name");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：name");
            return false;
        }
        else
            info.name = itr->value.GetString();

        itr = doc.FindMember("desc");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：desc");
            return false;
        }
        else
            info.desc = itr->value.GetString();

        itr = doc.FindMember("type");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：type");
            return false;
        }
        else
            info.type = itr->value.GetInt();

        itr = doc.FindMember("publicKey");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：publicKey");
            return false;
        }
        else{
            std::string pk = itr->value.GetString();
            std::transform(pk.begin(), pk.end(), pk.begin(), ::tolower);
            info.publicKey = pk;
        }

        itr = doc.FindMember("externalIP");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：externalIP");
            return false;
        }
        else
            info.externalIP = itr->value.GetString();

        itr = doc.FindMember("internalIP");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：internalIP");
            return false;
        }
        else
            info.internalIP = itr->value.GetString();

        itr = doc.FindMember("rpcPort");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：rpcPort");
            return false;
        }
        else
            info.rpcPort = itr->value.GetInt();

        itr = doc.FindMember("p2pPort");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：p2pPort");
            return false;
        }
        else
            info.p2pPort = itr->value.GetInt();

        itr = doc.FindMember("root");
        if (itr == doc.MemberEnd())
        {
            bcwasm::println("can not find member：root");
            return false;
        }
        else
            info.root = itr->value.GetBool();

        info.status = 0;
        
        return true;
    }

    //序列化注册信息
    // list repuired
    char *serializeRegisterInfo(int code, const char *message, std::vector<RegisterInfo> list) const
    {
        rapidjson::Document jsonDoc;                                            //生成一个dom元素Document
        rapidjson::Document::AllocatorType &allocator = jsonDoc.GetAllocator(); //获取分配器
        jsonDoc.SetObject();                                                    //将当前的Document设置为一个object，也就是说，整个Document是一个Object类型的dom元素

        //添加属性
        jsonDoc.AddMember("code", code, allocator);

        rapidjson::Value msg(message, allocator);
        jsonDoc.AddMember("msg", msg, allocator);

        //生成一个object数组
        rapidjson::Value registerArray(rapidjson::kArrayType);

        for (std::vector<RegisterInfo>::iterator it = list.begin(); it != list.end(); ++it)
        {
            rapidjson::Value regist(rapidjson::kObjectType);

            rapidjson::Value name(it->name.c_str(), allocator);
            rapidjson::Value owner(it->owner.c_str(), allocator);
            rapidjson::Value desc(it->desc.c_str(), allocator);
            rapidjson::Value publicKey(it->publicKey.c_str(), allocator);
            rapidjson::Value externalIP(it->externalIP.c_str(), allocator);
            rapidjson::Value internalIP(it->internalIP.c_str(), allocator);
            rapidjson::Value approver(it->approver.c_str(), allocator);

            regist.AddMember("name", name, allocator);
            regist.AddMember("owner", owner, allocator);
            regist.AddMember("desc", desc, allocator);
            regist.AddMember("type", it->type, allocator);
            regist.AddMember("publicKey", publicKey, allocator);
            regist.AddMember("externalIP", externalIP, allocator);
            regist.AddMember("internalIP", internalIP, allocator);
            regist.AddMember("rpcPort", it->rpcPort, allocator);
            regist.AddMember("p2pPort", it->p2pPort, allocator);
            regist.AddMember("status", it->status, allocator);
            regist.AddMember("root", it->root, allocator);
            regist.AddMember("approver", approver, allocator);
            regist.AddMember("registerTime", it->registerTime, allocator);

            registerArray.PushBack(regist, allocator); //添加到数组
        }

        jsonDoc.AddMember("data", registerArray, allocator);

        //生成字符串
        rapidjson::StringBuffer buffer;
        rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
        jsonDoc.Accept(writer);

        std::string strJson = buffer.GetString();
        bcwasm::println("serializeRegisterInfo info ", strJson.c_str());
        RETURN_CHARARRAY(strJson.c_str(), strJson.size() + 1);
    }

    //序列化节点信息
    // info repuired
    char *serializeNodeInfo(RegisterInfo info) const
    {
        rapidjson::Document jsonDoc;                                            //生成一个dom元素Document
        rapidjson::Document::AllocatorType &allocator = jsonDoc.GetAllocator(); //获取分配器
        jsonDoc.SetObject();                                                    //将当前的Document设置为一个object，也就是说，整个Document是一个Object类型的dom元素

        rapidjson::Value name(info.name.c_str(), allocator);
        rapidjson::Value owner(info.owner.c_str(), allocator);
        rapidjson::Value desc(info.desc.c_str(), allocator);
        rapidjson::Value publicKey(info.publicKey.c_str(), allocator);
        rapidjson::Value externalIP(info.externalIP.c_str(), allocator);
        rapidjson::Value internalIP(info.internalIP.c_str(), allocator);
        rapidjson::Value approver(info.approver.c_str(), allocator);

        jsonDoc.AddMember("name", name, allocator);
        jsonDoc.AddMember("owner", owner, allocator);
        jsonDoc.AddMember("approver", approver, allocator);
        jsonDoc.AddMember("desc", desc, allocator);
        jsonDoc.AddMember("type", info.type, allocator);
        jsonDoc.AddMember("publicKey", publicKey, allocator);
        jsonDoc.AddMember("externalIP", externalIP, allocator);
        jsonDoc.AddMember("internalIP", internalIP, allocator);
        jsonDoc.AddMember("rpcPort", info.rpcPort, allocator);
        jsonDoc.AddMember("p2pPort", info.p2pPort, allocator);
        jsonDoc.AddMember("status", info.status, allocator);
        jsonDoc.AddMember("root", info.root, allocator);

        //生成字符串
        rapidjson::StringBuffer buffer;
        rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
        jsonDoc.Accept(writer);

        std::string strJson = buffer.GetString();
        RETURN_CHARARRAY(strJson.c_str(), strJson.size() + 1);
    }
};
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::NodeRegister, registerNode)
BCWASM_ABI(SystemContract::NodeRegister, approve)
BCWASM_ABI(SystemContract::NodeRegister, getRegisterInfoByStatus)
BCWASM_ABI(SystemContract::NodeRegister, getRegisterInfoByPublicKey)
BCWASM_ABI(SystemContract::NodeRegister, getRegisterInfoByOwnerAddress)
//bcwasm autogen begin
extern "C" { 
int registerNode(const char * nodeJson) {
SystemContract::NodeRegister NodeRegister_bcwasm;
return NodeRegister_bcwasm.registerNode(nodeJson);
}
int approve(char * publicKey,int status) {
SystemContract::NodeRegister NodeRegister_bcwasm;
return NodeRegister_bcwasm.approve(publicKey,status);
}
const char * getRegisterInfoByStatus(int status,int pageNum,int pageSize) {
SystemContract::NodeRegister NodeRegister_bcwasm;
return NodeRegister_bcwasm.getRegisterInfoByStatus(status,pageNum,pageSize);
}
const char * getRegisterInfoByPublicKey(char * publicKey) {
SystemContract::NodeRegister NodeRegister_bcwasm;
return NodeRegister_bcwasm.getRegisterInfoByPublicKey(publicKey);
}
const char * getRegisterInfoByOwnerAddress(char * owner) {
SystemContract::NodeRegister NodeRegister_bcwasm;
return NodeRegister_bcwasm.getRegisterInfoByOwnerAddress(owner);
}
void init() {
SystemContract::NodeRegister NodeRegister_bcwasm;
NodeRegister_bcwasm.init();
}

}
//bcwasm autogen end