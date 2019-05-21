#!/bin/bash

#初始化链（删除４个节点保存的区块，重新生成创世区块）
./init_one.sh &&

#启动主节点
./start_one.sh &&

#启动从节点Ａ，Ｂ，Ｃ
# ./start-ABC.sh &&

#等待节点启动
sleep 3 &&

#解锁根用户
./unlock_root.sh &&

#部署系统合约
./deploy-system-contracts.sh