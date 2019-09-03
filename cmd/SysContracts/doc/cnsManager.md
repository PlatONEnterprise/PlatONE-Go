## 合约管理合约 RegisterManager

各个系统合约在本合约注册，用一个map保存各个合约的地址、合约名称、版本信息。

### 1. state变量

```
struct ContractInfo{
    string contract_name
    string contract_version
    address contract_address // 合约地址
    address origin // 创建者地址 暂保留，具体再讨论
    int64 create_time
    bool enabled 
}

map [] //保存所有合约信息的map；

```

### 2. 注册合约 | cnsRegisterFromInit(name, version)

将已经部署的合约注册到本合约的中存储。保存所有的历史版本。

* 必须从init()函数中调用注册，如果从合约的非init方法中调用，则注册失败。
* inputs: 

```
name：
	1. 合约名，需要字母数字或者下划线打头
	2. 合约名的所有者为第一次注册合约的部署者
version：
	1. 合约版本：需满足[num].[num].[num].[num]的格式，如1.0.0.0
	2. 注册版本必须逐步递增，不能回退注册
```

* output:
```
0：注册成功
1：注册失败
```

### 3. 注册合约 | cnsRegister(name, version, address)

将已经部署的合约注册到本合约的中存储。保存所有的历史版本。

- 只有合约的所有者才能调用该接口

* 注册合约时，既可以由用户或者合约的对外接口调用，还可以支持合约在自己的init()函数中调用。未来将移除对init()方法中调用的支持。即如果需要在init()方法中调用统一使用cnsRegisterFromInit()接口。

* inputs: 

```
name：
	1. 合约名，需要字母数字或者下划线打头
	2. 合约名的所有者为第一次注册合约的部署者
version：
	1. 合约版本：需满足[num].[num].[num].[num]的格式，如1.0.0.0
	2. 注册版本必须逐步递增，不能回退注册

address：合约地址必须符合以太坊合约地址标准格式，“0x”前缀可写可不写
```

* output:
```
0：注册成功
1：注册失败
```

### 4. 注销特定合约 | unregister(name, version)

- 只有合约的owner才能调用该接口

* inputs: 

```
name：合约名，需要字母数字或者下划线打头
version：合约版本：需满足[num].[num].[num].[num]的格式，如1.0.0.0
```

* output:

```
0：注销成功
1：注销失败
```

### 5. 获取合约地址 | getContractAddress(name, version)
- version为“latest”是获取最新版本。

* inputs:

```
name：合约名，需要字母数字或者下划线打头
version：合约版本：需满足[num].[num].[num].[num]的格式，如1.0.0.0
```

* output: 

```
address：标准以太坊地址（含“0x”前缀）
```

### 6. 获取已注册合约 | getRegisteredContracts(pageNum, pageSize)

- 两个参数必须为非负数

* inputs: 

```  
pageNum：页面编号（从0开始）
pageSize：每页显示条目数
```

* output:

```
合约信息的json字符串
```

### 7. 获取某人已注册合约 | getRegisteredContracts(address, pageNum, pageSize)

* inputs: 

```
address：合约地址必须符合以太坊合约地址标准格式，“0x”前缀可写可不写
pageNum：页面编号
pageSize：每页显示条目数
```

* output: 

```
合约信息的json字符串
```

### 8. 是否已注册 | ifRegisteredByName(name)

* inputs:

```  
name：合约名，需要字母数字或者下划线打头
```

* output: 

```
0：未注册（包含已失效）
1：已注册
```

### 9. 是否已注册 | ifRegisteredByAddress(address)

* inputs: 

```
address：合约地址必须符合以太坊合约地址标准格式，“0x”前缀可写可不写
```

* output: 

```
0：未注册（包含已失效）
1：已注册
```

### 10. 查询历史合约(包含已注销合约) | getHistoryContractsByName(name)
* inputs:

```
name：合约名，需要字母数字或者下划线打头
```

* output:

```
合约信息的json字符串
```