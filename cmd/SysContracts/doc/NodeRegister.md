## 节点申请合约NodeRegister

### RegisterInfo数据结构

```c++
struct RegisterInfo{
    string      name			//节点名称
    address     owner           //申请者的地址
    string      desc			//节点描述
    int         type	        //1:共识节点；0:观察者节点
    string      publicKey 		//节点公钥
    string      externalIP    	//外网 IP
    string      internalIP		//内网 IP
    int         rpcPort			//rpc 端口
    int         p2pPort 		//p2p端口
    int         status          //0:未审核；1:申请通过；2:拒绝申请
    address     approveor       //审核人的地址
    int64_t     registerTime	//申请时间	
}
```

### 合约接口说明

#### register(RegisterInfo)

注册节点信息

- inputs: 

```json
RegisterInfo是Json字符串，包含以下内容：
{
    "name":"node1",
    "desc":"i love this world",
    "type":1,
    "publicKey":"acb2281452fb9fc25d40113fb6afe82b498361de0eea6449c2502",
    "externalIP":"127.0.0.1",
    "internalIP":"127.0.0.1",
    "rpcPort":4789,
    "p2pPort":14789
}
· 只能由chainCreater、chainAdmin、nodeAdmin调用
· name：只可包含字母、数字、下划线、连字符，审核通过的name不可以重复
· desc：长度限制为4-1000
· type：必须是0或者1，0表示观察者节点
· publicKey：长度限制为4-1000，publicKey不能重复申请
· IP和端口：任意组合审核通过后，不能重复申请
```

- output: int

#### approve(publicKey, status)

审核节点申请信息

- inputs: 

```
· publicKey:不能为空，如果没申请过则无效
· status：必须是1或者2，1表示通过，2表示拒绝
· 只能由chainCreater、chainAdmin调用
· 如果状态是已审核，不能重新审核
· 审核通过的name、IP+Port不可以再次审核通过
· 审核通过（status = 1）后，调用NodeManager合约的add()方法把节点信息放入节点管理合约
```

- output: int

#### getRegisterInfoByStatus(status, pageNum, pageSize)

根据状态获取申请信息，分页输出

- inputs: status, pageNum, pageSize

```json
status:查询状态，0申请未处理、1申请通过、2申请拒绝
pageNum：页面编号（从0开始）
pageSize：每页显示条目数
```

- output: `[]RegisterInfo`

```json
返回Json字符串 
{
	"code":0,
    "msg":"",
    "data":{
    	"name":"node1",
    	"owner":"0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
    	"desc":"i love this world",
    	"type":1,
        "publicKey":"acb2281452fb9fc25d40113fb6afe82b498361de0eea6449c2502",
    	"externalIP":"127.0.0.1",
    	"internalIP":"127.0.0.1",
    	"rpcPort":4789,
    	"p2pPort":14789,
    	"status":0,
        "approveor":"0x4FCd6fe35f0612C7866943cb66C1d93eb0746bcC",
        "registerTime":23489723
    }
}
```

#### getRegisterInfoByNodeAddress(publicKey)

根据公钥获取申请信息

- inputs: publicKey

```
· publicKey：申请过的公钥
```

- output: `[]RegisterInfo`

```
返回Json字符串
```

#### getRegisterInfoByOwnerAddress(ownerAddress)

根据申请信息owner获取申请信息

- inputs: ownerAddress

```
· ownerAddress：申请者地址
```

- output: `[]RegisterInfo`

```
返回Json字符串
```
