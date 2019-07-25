## 4. 平台用户申请合约 UserRegisterManager

申请用户指的是，用户将个人的账户地址、用户名称、电话、邮箱、个人说明，向合约发起申请，合约保存信息。

## 4.1. 数据结构说明

```
struct RegisterInfo 
{
		std::string user_address;     // 用户账户地址
		std::string name;             // 用户名称
		std::string mobile;           // 手机号
		std::string email;            // 邮箱
		std::string remark;           // 个人说明
		unsigned int user_state;      // 平台用户状态：1审核中，2已激活，3已拒绝
		std::string auditor_address;  // 审核人的地址
		std::vector<string> roles;   // 角色请求列表
};
```

## 4.2. 注册的合约名称
* 注册的合约名称为：__sys_UserRegisterManager

## 4.3. 合约接口说明
### (1) int registerUser(const char* registJson)
功能：添加用户申请信息。  

入参：

* registJson 用户申请信息JSON结构体字符串   

出参：

* int 添加用户申请成功返回0，失败返回-1.

用户信息json结构体字符串如下：

```
{
	"address":"0x33d253386582f38c66cb5819bfbdaad0910339b3",
	"name":"xiaoluo",
	"mobile":"13111111111",
	"email":"luodahui@qq.com",
	"roles":["chainAdmin","nodeAdmin"],
	"remark":"平台用户申请"
}
```

### (2) int approve(const char* userAddress, int auditStatus)
功能：审核用户申请信息

入参：

* userAddress 用户地址   
* auditStatus 审核状态： 2 激活，3 拒绝

出参：

* int 审核成功返回0，失败返回-1.

### (3) const char* getAccountByAddress(const char* address) const
功能：根据地址获取用户申请信息

入参：

* address 用户地址   

出参：

* const char* 返回json格式字符串，格式如下：
```
{
	"code":0,
	"msg":"succeed",
	"data":{
		....
	}
}
```

如果用户申请不存在，返回错误信息：

```
{
	"code":1,
	"msg":"The user does not exist in the userRegisterManager",
	"data":""
}
```

### (4) const char* getAccountByUsername(const char* UserName) const
功能：根据用户名获取用户申请信息

入参：

* UserName 用户名  

出参：

* const char* 返回json格式字符串，格式如下：
```
{
	"code":0,
	"msg":"succeed",
	"data":{
		....
	}
}
```

如果用户不存在，返回错误信息：

```
{
	"code":1,
	"msg":"The user does not exist in the userRegisterManager",
	"data":""
}
```

### (5) const char* getAccountsByStatus(int pageNum, int pageSize, int accountStatus) const
功能：分页查询返回某个status的所有申请用户信息

入参：

* pageNum 页码 
* pageSize 每页的数据条数
* accountStatus 用户状态：1审核中，2已激活，3已拒绝

出参：

* const char* 返回json格式字符串，格式如下：
```
{
	"code":0,
	"msg":"succeed",
	"data":[{
		....
	}]
}
```

如果pageNum和pageSize不正确，如：
```
unsigned int startIndex = pageNum * pageSize;
unsigned int endIndex = startIndex + pageSize - 1;
if (startIndex >= size) 
{
	code = "1";
	message = "Adjust pageNum and pageSize";
}
```

返回错误信息：
```
{
	"code":1,
	"msg":"Adjust pageNum and pageSize",
	"data":""
}
```


如果用户状态accountStatus不存在，返回错误信息：
```
{
	"code":2,
	"msg":"The user status for the query does not exist in the userRegisterManager",
	"data":""
}
```

如果申请用户信息不存在，返回错误信息：
```
{
	"code":3,
	"msg":"user information is empty in the userRegisterManager",
	"data":""
}
```

