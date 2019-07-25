#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>
#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>
#include "../common/util.hpp"

using namespace rapidjson;

/*
struct NodeInfo
{
    string name;       // 节点名字
    address owner;     // 申请者的地址
    string desc;       // 节点描述
    int type;          // 0:观察者节点；1:共识节点
    string publicKey;  // 节点公钥
    string externalIP; // 外网 IP
    string internalIP; // 内网 IP
    int rpcPort;       // rpc 通讯端口
    int p2pPort;       // p2p 通讯端口
    int status;        // 1:正常；3：删除
    address approveor; // 审核人的地址
    int delayNum;      // 共识节点延迟设置的区块高度 (可选, 默认实时设置)
}
*/

namespace systemContract
{
    class NodeManager : public bcwasm::Contract
    {
    public:
        NodeManager() {}
        void init()
        {
            bcwasm::DeployedContract regManagerContract("0x0000000000000000000000000000000000000011");
            regManagerContract.call("cnsRegisterFromInit", "__sys_NodeManager", "1.0.0.0");

            bcwasm::println("NodeManager init success...");
        }

        /// 定义Event.
        BCWASM_EVENT(Notify, uint64_t, const char *)

    public:
        enum NodeAddCode
        {
            SUCCESS=0,
            BAD_PARAMETER,
            NO_PERMISSION
        };

        int add(const char *nodeJsonStr)
        {
            bcwasm::println("NodeManager add");
            Document node;
            node.Parse(nodeJsonStr);

            string origin = bcwasm::origin().toString();
            util::formatAddress(origin);

            // 必要的key
            std::vector<std::string> needfulKeys{"name", "owner", "desc", "type", "publicKey", "externalIP", "internalIP", "rpcPort", "p2pPort", "status"};
            uint64_t code = SUCCESS;
            std::string msg = "add node info success";

            std::vector<std::string> numType = {"type", "rpcPort", "p2pPort", "status"};                                   // 此字段对应的类型应为 数字类型
            std::vector<std::string> stringType = {"owner", "desc", "publicKey", "externalIP", "internalIP", "approveor"};

            do
            {
                bcwasm::DeployedContract cnsContract("0x0000000000000000000000000000000000000011");
                std::string urmAddress = cnsContract.callString("getContractAddress", "__sys_RoleManager", "latest");

                bcwasm::DeployedContract urmContract(urmAddress);
                std::string roleStr = urmContract.callString("getRolesByAddress", origin);

                bcwasm::println("urmAddress address ", urmAddress.c_str(), " roleStr ", roleStr.c_str());

                if (std::string::npos == roleStr.find("chainCreator") && std::string::npos == roleStr.find("chainAdmin") && std::string::npos == roleStr.find("nodeAdmin"))
                {
                    code = NO_PERMISSION;
                    msg = origin + " has no permission to add node info, role ret : " + roleStr;
                    break;
                }

                if (!node.IsObject())
                {
                    code = BAD_PARAMETER;
                    msg = "Node info json str is not a object";
                    break;
                }

                if (!checkType(node, numType, stringType))
                {
                    code = BAD_PARAMETER;
                    msg = "Node info json str data type is wrong";
                    break;
                }

                std::vector<std::string> lossKeys; // 一次性告诉插入数据的人少了哪些东西
                for (const auto &key : needfulKeys)
                {
                    if (!node.HasMember(key.c_str()))
                    {
                        lossKeys.push_back(key);
                    }
                }

                if (lossKeys.size() != 0)
                {
                    msg = "Node info json str has not member: ";
                    for (const auto &key : lossKeys)
                    {
                        msg += key + std::string(",");
                    }
                    msg.pop_back();
                    code = BAD_PARAMETER;
                    break;
                }

                // 重复性检查
                std::string publicKey = std::string(node["publicKey"].GetString());
                std::string name = std::string(node["name"].GetString());
                int type = node["type"].GetInt();
                std::string condition = "";

                // 公钥必须唯一
                condition = "{\"publicKey\":\"" + publicKey + "\"}";
                if (nodesNum(condition.c_str()) > 0)
                {
                    msg = "publicKey not unique";
                    code = BAD_PARAMETER;
                    break;
                }

                // 节点名字不能与非删除节点重复
                condition = "{\"name\":\"" + name + "\",\"status\":1}";
                int num1 = nodesNum(condition.c_str());
                if (num1 > 0)
                {
                    msg = "name not unique";
                    code = BAD_PARAMETER;
                    break;
                }

                // 添加节点时类型必须是观察者节点
                if (type != 0)
                {
                    msg = "join node type must be observer";
                    code = BAD_PARAMETER;
                    break;
                }

            } while (false);

            if (code == SUCCESS)
            {
                std::string nodesName;
                bcwasm::getState(NodeManager::dbKey, nodesName);
                nodesName += nodesName.size() == 0 ? std::string(node["name"].GetString()) : (std::string("|") + std::string(node["name"].GetString()));
                bcwasm::setState(NodeManager::dbKey, nodesName);

                bcwasm::setState(std::string(node["name"].GetString()), std::string(nodeJsonStr));
            }

            msg += std::string(". node info: ") + nodeJsonStr;
            BCWASM_EMIT_EVENT(Notify, code, msg.c_str());
            return code;
        }

        const char *getAllNodes() const
        {
            std::string ret;
            ret += R"({"code":0,)";
            ret += R"("msg":"success",)";
            ret += R"("data":[)";

            std::vector<std::string> names = getNodesName();
            for (auto &name : names)
            {
                std::string nodeStr;
                bcwasm::getState(name, nodeStr);
                ret += (nodeStr + ",");
            }

            if (names.size() > 0)
            {
                ret.pop_back();
            }
            ret += "]}";
            return util::makeReturnedStr(ret);
        }

        int validJoinNode(const char *publicKey) const
        {
            bcwasm::println("NodeManager publicKeyExist...", publicKey);
            std::string condition1 = "{\"publicKey\":\"" + std::string(publicKey) + "\",\"status\":1}";
            int ret = nodesNum(condition1.c_str());
            return ret;
        }

        // 根据条件查看符合节点的个数
        int nodesNum(const char *nodeJsonStr) const
        {
            const char *ret = getNodes(nodeJsonStr);
            Document nodes;
            nodes.Parse(ret);
            bcwasm::println("nodesNum = ", nodeJsonStr, ret);
            int size = nodes["data"].Size();
            return size;
        }

        // 根据条件返回符合条件的节点信息
        const char *getNodes(const char *nodeJsonStr) const
        {
            std::string ret;
            ret += R"({ "code":0,)";
            ret += R"("msg":"success",)";
            ret += R"("data":[)";

            size_t findSize = 0;

            Document inNode;
            inNode.Parse(nodeJsonStr);

            if (inNode.IsObject())
            {
                std::vector<std::string> names = getNodesName();
                for (auto &name : names)
                {
                    bool find = true;

                    std::string nodeStr;
                    bcwasm::getState(name, nodeStr);

                    Document curNode;
                    curNode.Parse(nodeStr.c_str());

                    for (Value::ConstMemberIterator itr = inNode.MemberBegin(); itr != inNode.MemberEnd(); ++itr)
                    {
                        Value::ConstMemberIterator curItr = curNode.FindMember(itr->name);
                        if (curItr != curNode.MemberEnd() && itr->value != curItr->value)
                        {
                            find = false;
                            break;
                        }
                    }

                    if (find)
                    {
                        findSize += 1;
                        ret += (nodeStr + ",");
                    }
                }

                if (findSize > 0)
                {
                    ret.pop_back();
                }
            }

            ret += "]}";

            bcwasm::println("in json info = ", nodeJsonStr, " , ret =  ", ret.c_str());
            return util::makeReturnedStr(ret);
        }

        int update(const char *name, const char *nodeJsonStr)
        {
            int updateCount = 0;
            Document inNode;
            inNode.Parse(nodeJsonStr);
            if (!inNode.IsObject())
            {
                BCWASM_EMIT_EVENT(Notify, 0, (std::string(nodeJsonStr) + std::string(" is not a json object")).c_str());
                return updateCount;
            }

            if (false == checkCallerPermission()){
                        BCWASM_EMIT_EVENT(Notify, 0, (std::string(nodeJsonStr) + std::string(" no permission")).c_str());
                        return updateCount;
            }
            std::string nodeStr;
            bcwasm::getState(std::string(name), nodeStr);
            if (nodeStr.empty())
            {
                BCWASM_EMIT_EVENT(Notify, 0, ("node " + std::string(name) + " has not found").c_str());
                return updateCount;
            }

            Document curNode;
            curNode.Parse(nodeStr.c_str());
            bcwasm::println("NodeManager update:", name, nodeJsonStr, nodeStr);

            for (Value::ConstMemberIterator itr = inNode.MemberBegin(); itr != inNode.MemberEnd(); ++itr)
            {
                std::string key = std::string(itr->name.GetString());
                // 更新节点信息只能是：desc, type, status, delayNum
                if (key == "desc" || key == "type" || (key == "status" && curNode["status"].GetInt() != 3) || key == "delayNum")
                {
                    curNode.RemoveMember(itr->name.GetString());
                    curNode.AddMember(rapidjson::StringRef(itr->name.GetString()), inNode[itr->name.GetString()], curNode.GetAllocator());
                    bcwasm::println("NodeManager update key:", key);
                    BCWASM_EMIT_EVENT(Notify, 0, (std::string("NodeManager update key:") + key).c_str());
                    updateCount++;
                }
            }

            StringBuffer buffer;
            Writer<StringBuffer> writer(buffer);
            curNode.Accept(writer);
            const char *output = buffer.GetString();

            curNode.Parse(std::string(buffer.GetString()).c_str());
            std::vector<std::string> numType = {"type", "rpcPort", "p2pPort", "status", "delayNum"};                                   // 此字段对应的类型应为 数字类型
            std::vector<std::string> stringType = {"owner", "desc", "publicKey", "externalIP", "internalIP", "approveor"};
            if (!checkType(curNode, numType, stringType))
            {
                BCWASM_EMIT_EVENT(Notify, 0, ("node " + std::string(name) + "Node info json str data type is wrong").c_str());
                return updateCount;
            }

            bcwasm::setState(std::string(name), std::string(buffer.GetString()));
            return updateCount;
        }

        const char *getEnodeNodes(int deleted) const
        {
            bcwasm::println("NodeManager getEnodeNodes...");
            std::string ret;

            std::vector<std::string> names = getNodesName();
            for (auto &name : names)
            {
                std::string nodeStr;
                bcwasm::getState(name, nodeStr);
                Document node;
                node.Parse(nodeStr.c_str());

                int status = node["status"].GetInt();
                if (deleted && status <= 2)
                {
                    continue;
                }

                if (!deleted && status == 3)
                {
                    continue;
                }

                std::string publicKey = std::string(node["publicKey"].GetString());
                std::string internalIP = std::string(node["internalIP"].GetString());
                std::string p2pPort = std::to_string(node["p2pPort"].GetInt());
                ret += std::string("enode://") + publicKey + std::string("@") + internalIP + std::string(":") + p2pPort + std::string("|");
            }

            if (!ret.empty())
            {
                ret.pop_back();
            }

            return util::makeReturnedStr(ret);
        }

        const char *getNormalEnodeNodes() const
        {
            bcwasm::println("NodeManager getEnodeNodes...");
            return getEnodeNodes(0);
        }

        const char *getDeletedEnodeNodes() const
        {
            return getEnodeNodes(1);
        }

        static std::string dbKey;

    private:
        bool checkType(Document &node, std::vector<std::string> &numType, std::vector<std::string> &stringType) const
        {
            bool ret = true;

            for (Value::ConstMemberIterator itr = node.MemberBegin(); itr != node.MemberEnd(); ++itr)
            {
                std::string key = std::string(itr->name.GetString());

                auto vitr = std::find(numType.begin(), numType.end(), key);
                if (vitr != numType.end() && itr->value.GetType() != kNumberType)
                {
                    ret = false;
                }

                vitr = std::find(stringType.begin(), stringType.end(), key);
                if (vitr != stringType.end() && itr->value.GetType() != kStringType)
                {
                    ret = false;
                }
            }

            return ret;
        }
        std::vector<std::string> getNodesName() const
        {
            std::vector<std::string> v;
            std::string s;
            bcwasm::getState(NodeManager::dbKey, s);
            std::string c = "|";

            std::string::size_type pos1, pos2;
            pos2 = s.find(c);
            pos1 = 0;
            while (std::string::npos != pos2)
            {
                v.push_back(s.substr(pos1, pos2 - pos1));

                pos1 = pos2 + c.size();
                pos2 = s.find(c, pos1);
            }
            if (pos1 != s.length())
                v.push_back(s.substr(pos1));
            return v;
        }
        bool checkCallerPermission(){

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
                //bcwasm::println("用户状态出错");
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
                if (str == util::CHAIN_CREATOR || str ==util::CHAIN_ADMIN || util::NODE_ADMIN){
                    return true;
                }

            }
            return false;
        }

    };
    std::string NodeManager::dbKey = "nodesName";

} // namespace systemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(systemContract::NodeManager, add)
BCWASM_ABI(systemContract::NodeManager, getAllNodes)
BCWASM_ABI(systemContract::NodeManager, getNodes)
BCWASM_ABI(systemContract::NodeManager, nodesNum)
BCWASM_ABI(systemContract::NodeManager, update)
BCWASM_ABI(systemContract::NodeManager, validJoinNode)
BCWASM_ABI(systemContract::NodeManager, getEnodeNodes)
BCWASM_ABI(systemContract::NodeManager, getNormalEnodeNodes)
BCWASM_ABI(systemContract::NodeManager, getDeletedEnodeNodes)
//bcwasm autogen begin
extern "C" { 
int add(const char * nodeJsonStr) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.add(nodeJsonStr);
}
const char * getAllNodes() {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.getAllNodes();
}
int validJoinNode(const char * publicKey) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.validJoinNode(publicKey);
}
int nodesNum(const char * nodeJsonStr) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.nodesNum(nodeJsonStr);
}
const char * getNodes(const char * nodeJsonStr) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.getNodes(nodeJsonStr);
}
int update(const char * name,const char * nodeJsonStr) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.update(name,nodeJsonStr);
}
const char * getEnodeNodes(int deleted) {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.getEnodeNodes(deleted);
}
const char * getNormalEnodeNodes() {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.getNormalEnodeNodes();
}
const char * getDeletedEnodeNodes() {
systemContract::NodeManager NodeManager_bcwasm;
return NodeManager_bcwasm.getDeletedEnodeNodes();
}
void init() {
systemContract::NodeManager NodeManager_bcwasm;
NodeManager_bcwasm.init();
}

}
//bcwasm autogen end