### 合约测试脚本使用

#### 0. 准备工作：

通过快速起链脚本 启动测试链：

```
./platonectl one
```



#### 1. 步骤：

1. 进入到某个合约命名的文件夹 如：

   `cd template`

2. 编辑`cfg`文件：

   ```shell
   CTOOL_JSON=../../../release/linux/conf/ctool.json
   ABI_JSON=../../../release/linux/conf/contracts/roleRegister.cpp.abi.json
   ADDR=0x7843befb3bbf1e5d625d876b5382f4f00fdcfd87
   CTOOL_BIN=../../../release/linux/bin/ctool
   WASM_LOG=../../../release/linux/data/node-0/logs/wasm_log/*
   ```

   1）指定正确的文件路径（如在当前路径下测试，可保持默认路径不变）

   2）其中`ADDR`变量指示当前要测试的合约地址



3. 为合约每个接口编写入口脚本，如

   ```shell
   #!/bin/bash
   cfg='cfg'
   config=`cat $cfg | grep CTOOL_JSON | awk -F'=' '{print $2}'`
   abi=`cat $cfg | grep ABI_JSON | awk -F'=' '{print $2}'`
   addr=`cat $cfg | grep ADDR | awk -F'=' '{print $2}'`
   ctool=`cat $cfg | grep CTOOL_BIN | awk -F'=' '{print $2}'`
   
   roles_string=$1
   
   $ctool invoke --config $config --addr $addr --abi $abi --func registerRole --param $roles_string
   echo "registerRole"
   echo roles_string = $roles_string
   ```

   （前6行不用变，保留修改`invoke`行，及其入参即可）

4. 编辑runTest脚本，移动到`# START TO MODIFY HERE`行，开始编辑用例，一个用例示例如下：

   ```shell
   params=(case1 case2p1|case2p2) # 指示合约方法的入参，以空格分割。（如果一次合约调用有多个参数，多个参数以|分割）
   expect_code=(0 1) # 指示期待的执行结果，以空格分割。（默认0为正确，非0为错误）
   cmd=./registerRole.sh # 执行的合约方法
   runTest
   ```

5.  执行`./runTest.sh`，输出形式如下：

   ```shell
   START RUNNING CASE FOR ./registerRole.sh
   
   Case 0:
   input param is:  ["contractAdmin"]
   wasm.log: 
   INFO [09-05|11:59:23.221|core/vm/logger.go:291]                         [RoleRegister] account code:  0
   OK: [RoleRegister] [registerRole] Register success.
    RoutineID=83
   
   ---------------------------------
   
   
   START RUNNING CASE FOR ./getRegisterInfosByStatus.sh
   
   Case 0:
   input param is:  1 0 10
   result: {"code":0,"msg":"ok","data":[{"userAddress":"0xe57712be2c5d8cce1e7679156e18d0bebcaaa0a4","userName":"root","roleRequireStatus":1,"requireRoles":["contractAdmin"],"approver":""}]}
    [ OK ] 
   Case 1:
   input param is:  2 0 10
   result: {"code":1,"msg":"not found","data":[]}
    [ OK ] 
   Case 2:
   input param is:  0 0 10
   result: {"code":1,"msg":"not found","data":[]}
    [ OK ] 
   Case 3:
   input param is:  1 0 0
   result: {"code":0,"msg":"ok","data":[{"userAddress":"0xe57712be2c5d8cce1e7679156e18d0bebcaaa0a4","userName":"root","roleRequireStatus":1,"requireRoles":["contractAdmin"],"approver":""}]}
    [ OK ] 
   ---------------------------------
   
   ```

   

   
