## 6. 平台角色申请合约	 RoleRegister

用以申请角色、审批角色、以及查询角色申请信息。

### 6.1. state变量

```c++
    typedef struct RegisterUserInfo {
        std::string userAddress;       //账户地址
        std::string userName;            //用户名
        uint8_t roleRequireStatus;  //角色申请状态 1 申请中 2 已批准 3 已拒绝
        std::vector<std::string> requireRoles; //待申请角色列表
        std::string approveor;         //审核人的地址
        BCWASM_SERIALIZE(RegisterUserInfo, (userAddress)(userName)(roleRequireStatus)(requireRoles)(approveor));
    }RegisterUserInfo_st;
```

### 6.2. 注册角色 | registerRole(const char* roles)
记录用户申请信息,一个账户可以同时申请多个角色。

* inputs : const char* roles
  - eg: [ \\"chainAdmin\\",\\"nodeAdmin\\"]
* output: uint8_t
  - 0 注册成功
  - 1 地址状态不正确
  - 2 内部错误
  - 3 该地址已含被拒绝记录
  - 4 输入参数不正确，不是list类型
  - 5 调用者无权限调用本方法
  - 6 没有有效的申请信息
  - 7 无效的待申请角色名称
  

### 6.3. 审批角色 | approve(const char* address, uint8_t status)
只能审核比自己权限更低的角色申请。

* inputs: address, status
* output: uint8_t
  - 0 注册成功
  - 1 账号地址格式错误
  - 2 角色权限不足
  - 3 没有角色申请信息 
  - 4 内部错误
  - 5 不是合法的审批值
  - 6 当前是非审批状态 

* 行为
    1. 更改申请信息的status状态
    2. 管理员审核通过后，调用RoleManager的addRole()接口

### 6.4. 根据地址查询角色申请信息 | getRegisterInfoByAddress(const char* address)

* inputs: address
* output: {状态值, msg, `RegisterUserInfo`}

返回值格式如下：
```json
{
    code: 0,
    msg: 'ok',
    data: {
        "userAddress": "0x111222",
        "userName": "xiaoming",
        "roleRequireStatus": 1,
        "requireRoles":["Admin", "ContractCaller"],
        "approveor": "01xxx"
    }
}
```

### 6.5. 根据用户名查询角色申请信息 | getRegisterInfoByName(const char* name)
* inputs: name
* output: {状态值, msg, `RegisterUserInfo`}

  
返回值格式如下：
```json
{
    code: 0,
    msg: 'ok',
    data: {
        "userAddress": "0x111222",
        "userName": "xiaoming",
        "roleRequireStatus": 1,
        "requireRoles":["Admin", "ContractCaller"],
        "approveor": "01xxx"
    }
}
```

### 6.6.  根据申请状态查询角色申请信息 | getRegisterInfosByStatus(int status, int pageNum, int pageSize)

* inputs: status, pageNum, pageSize
* output: {状态值, msg, []`RegisterUserInfo`}
  
返回值格式如下：
```json
{
    code: 0,
    msg: 'done',
    data: [
        {
            "userAddress": "0x111222",
            "userName": "xiaoming",
            "roleRequireStatus": 1,
            "requireRoles":["Admin", "ContractCaller"],
            "approveor": "01xxx"
        },
        {
            "userAddress": "0x111222",
            "userName": "xiaohong",
            "roleRequireStatus": 1,
            "requireRoles":["Admin", "ContractCaller"],
            "approveor": "01xxx"
        }
    ]
}
```
