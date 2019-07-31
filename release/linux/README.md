###   Quick Start



#### 1. 节点操作

首先，进入`scripts`目录

1.1 启动单节点链：

   ```shell
   ./platonectl.sh one
   ```

1.2 在一台机器上快速启动四节点区块链:

   ```shell
   ./platonectl.sh four
   ```

1.3 查看链的运行状态：

```shell
./platonectl.sh status
```

1.4 停止某个节点，如停止节点0：

```shell
./platonectl.sh stop -n 0
```



#### 2.  RPC 默认端口

快速启动单节点的链：RPC端口为：6791

快速启动四节点的链，RPC端口分别为：6791-6794



#### 3. 默认账户

* keystore位置：

```
./data/node-[x]/keystore
```

* 账号地址：keystore文件中第一个address字段

* 默认账号解锁密码: **0**



#### 4. 日志

节点日志默认位置如下，主要打印节点运行信息：

```
./data/node-[x]/logs
```

wasm日志默认位置如下，主要打印合约调用中的输出：

```
./data/node-[x]/logs/wasm.log
```

