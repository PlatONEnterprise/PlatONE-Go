# 平台用户管理合约 UserManager

## 5.1. 数据结构说明

```
struct UserInfo
{
    string user_address;	// 用户地址
    string name;//用户名称 
    string email; //用户邮箱 
    string mobile; //手机号
    unsigned int status;        // 是否禁用用户：[0:可用; 1:禁用; 2:删除] 
};
```

## 5.2. 注册的合约名称
* 注册的合约名称为：__sys_UserManager

## 5.3. 合约接口说明
### (1) int addUser(const char* userJson)
功能：添加用户信息。  

入参：

* userJson 用户信息JSON结构体字符串   

出参：

* int 添加用户成功返回0，失败返回-1.

用户信息json结构体字符串如下：

```
{
"address":"0x33d253386582f38c66cb5819bfbdaad0910339b3",
"name":"xiaoluo",
"mobile":"13111111111",
"email":"luodahui@qq.com",
"status":0
}
```

### (2) int enable(const char* userAddr)
功能：将用户设置为可用状态。  

入参：

* userAddr 用户地址   

出参：

* int 设置成功返回0，失败返回-1.

### (3) int disable(const char* userAddr)
功能：将用户设置为禁用状态。  
入参：

* userAddr 用户地址   

出参：

* int 设置成功返回0，失败返回-1.

### (4) int delUser(const char* userAddr)
功能：将用户设置为删除状态 

入参：

* userAddr 用户地址   

出参：

* int 设置成功返回0，失败返回-1.

### (5) int update(const char* userAddr, const char* updateJson)
功能：更新用户信息，只能更新用户的email,mobile,status信息

入参：

* userAddr 用户地址   
* updateJson 更新信息的json结构体字符串

出参：

* int 更新成功返回0，失败返回-1.

### (6) const char* getAccountByAddress(const char* address) const
功能：根据地址获取用户信息

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

如果用户不存在，返回错误信息：

```
{
	"code":1,
	"msg":"The user does not exist in the useManager contract",
	"data":""
}
```

### (7) const char* getAccountByName(const char* name) const
功能：根据用户名获取用户信息

入参：

* name 用户名  

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
	"msg":"The user does not exist in the useManager contract",
	"data":""
}
```

### (8) int isValidUser(const char* userAddr) const
功能：判断是否是有效用户

入参：

* userAddr 用户地址  

出参：

* int 返回0为有效用户，-1为非有效用户

## 5.4. 限制
* 超管可编辑所有用户的基本信息；
* 普管可编辑除其他普管外的其他用户的基本信息；
* 用户可编辑自身基本信息；
* 账户地址和用户名不可更改，其他基本信息可编辑