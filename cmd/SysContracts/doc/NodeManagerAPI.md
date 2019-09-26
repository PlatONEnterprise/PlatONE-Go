# 系统合约之节点管理

## 数据结构说明

```c++
struct NodeInfo
{
    string name;       // 节点名字, 全网唯一，不能重复。所有接口均以此为主键。
    address owner;     // 申请者的地址
    string desc;       // 节点描述
    int type;          // 0:观察者节点；1:共识节点
    string publicKey;  // 节点公钥，全网唯一，不能重复
    string externalIP; // 外网 IP
    string internalIP; // 内网 IP
    int rpcPort;       // rpc 通讯端口
    int p2pPort;       // p2p 通讯端口
    int status;        // 1:正常；2：删除
    address approveor; // 审核人的地址
    int delayNum;      // 共识节点延迟设置的区块高度 (可选, 默认实时设置)
}
```

## 合约接口调用说明
### void add(const char *nodeJsonStr)
添加节点信息。  
入参：
* nodeJsonStr 节点信息JSON结构体字符串   

出参：
* void 

一个可能的入参如下所示，注意，传入的时候需要将此结构体转为字符串
```javascript
{
    "name": "node1",
    "owner": "0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
    "approveor": "0xdF6518e51e0d90A3CBa26e4AdFE74454E2D90BdE",
    "desc": "i love this world",
    "type": 1,
    "publicKey": "0x81ec63a2335c1f79244cbe485eb6bffef48cfd7df58b1009411c6114670eefd27da865914c70f7e49ceeb1002f1c24f4930975a2eb05cb5ac1373bed83a9932a",
    "externalIP": "127.0.0.1",
    "internalIP": "127.0.0.1",
    "rpcPort": 6789,
    "p2pPort": 16789,
    "status": 1
}
```

调用此接口过程中，对于一些常见错误会以事件的形式提示，供调用者调试。
*****
### const char *getAllNodes() const
返回所有插入的节点信息  
入参：
* 无

出参：
* nodes 所有节点信息

一个可能的所有节点返回信息如下所示，注意，本身是字符串，为了显示友好，已格式化为了json
```javascript
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "name": "node1",
      "owner": "0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
      "approveor": "0xdF6518e51e0d90A3CBa26e4AdFE74454E2D90BdE",
      "desc": "i love this world",
      "type": 1,
      "publicKey": "0x81ec63a2335c1f79244cbe485eb6bffef48cfd7df58b1009411c6114670eefd27da865914c70f7e49ceeb1002f1c24f4930975a2eb05cb5ac1373bed83a9932a",
      "externalIP": "127.0.0.1",
      "internalIP": "127.0.0.1",
      "rpcPort": 6789,
      "p2pPort": 16789,
      "status": 1
    },
    {
      "name": "node2",
      "owner": "0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
      "approveor": "0xdF6518e51e0d90A3CBa26e4AdFE74454E2D90BdE",
      "desc": "i love this world too!",
      "type": 0,
      "publicKey": "0x81ec63a2335c1f79244cbe485eb6bffef48cfd7df58b1009411c6114670eefd27da865914c70f7e49ceeb1002f1c24f4930975a2eb05cb5ac1373bed83a9932a",
      "externalIP": "10.10.8.160",
      "internalIP": "10.10.8.160",
      "rpcPort": 6789,
      "p2pPort": 16789,
      "status": 1
    }
  ]
}
```
*****
### const char *getNodes(const char *nodeJsonStr)
根据特定条件，返回符合条件的节点信息。比如，我需要查询节点的名字为"node1"的节点信息，那么传入**字符串** {"name":"node1"}；如果需要查询所有正常节点，那么传入字符串{"status":1}；如果我需要返回正常节点而且是共识节点，那么传入字符串{"status":1, "type":1}。总之，你可以根据节点信息可多个组合进行查询。  
入参：
* nodeJsonStr 条件信息，以json表示。

出参：
* nodes  符合查询条件的节点信息

假设我输入的查询信息为{"status":1, "type":1}，一个可能的所有节点返回信息如下所示，注意，本身是字符串，为了显示友好，已格式化为了json
```javascript
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "name": "node1",
      "owner": "0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
      "approveor": "0xdF6518e51e0d90A3CBa26e4AdFE74454E2D90BdE",
      "desc": "i love this world",
      "type": 1,
      "publicKey": "0x81ec63a2335c1f79244cbe485eb6bffef48cfd7df58b1009411c6114670eefd27da865914c70f7e49ceeb1002f1c24f4930975a2eb05cb5ac1373bed83a9932a",
      "externalIP": "127.0.0.1",
      "internalIP": "127.0.0.1",
      "rpcPort": 6789,
      "p2pPort": 16789,
      "status": 1
    }
  ]
}
```
*****
### void updata(const char *name, const char *nodeJsonStr)
更新节点信息，比如，我需要将节点的名字为"node1"的节点进行删除，那么第二个参数传入的**字符串** {"status":3}；如果我需要更新节点的内部IP以及rpc端口。那么第二个参数传入的字符串可能为：{"internalIP":"10.10.8.13", "rpcPort": 6788}。总之，你可以根据节点信息可多个组合进行更新。  
入参：

* name 节点名字
* nodeJsonStr 需要更新的信息，以json表示。
  

出参：

* void

