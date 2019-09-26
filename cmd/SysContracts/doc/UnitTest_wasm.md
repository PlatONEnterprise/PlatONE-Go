## 1. 编写wasm合约单元测试代码	
```c++
#include "../unittest.hpp"
TEST_CASE(hello, world) {
    ASSERT(true, "hello");
    int i = 1;
    int j = 2;
    ASSERT_NE(i, j, "xxxx");
}

UNITTEST_MAIN() {
    RUN_TEST(hello, world)
}
```

说明：
（1）. TEST_CASE里面写测试用例，RUN_TEST表示执行测试用例。
（2）. 将代码文件放在unitTest/testcase目录下。

具体样例参看合约代码，unitTest/testcase/unittest_userRegister.cpp

## 2. 编译wasm合约单元测试代码
执行script/autoUnitTest.sh脚本，在buildUnitTest/test/testcase目录下生成合约二进制文件。

## 3. 测试

使用cmd/wasm工具，执行测试用例。

命令：./wasm unittest --dir /SysContracts/buildUnitTest/unitTest/testcase -outdir /SysContracts/database --showLog 1

dir:合约二进制文件所在目录

outdir:单元测试输出目录,数据库文件保存在此目录的testdb下

showLog:是否显示日志标识，1：显示，默认不显示。