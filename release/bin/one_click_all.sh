#!/bin/bash

./ctl_all.sh stopall 
./ctl_all.sh cleanall
./ctl_all.sh cleanlog

#初始化链（删除４个节点保存的区块，重新生成创世区块）
./init_all.sh &&

#分发
./ctl_all.sh pkg &&

#启动主节点
./start_one.sh &&

#等待节点启动
sleep 3 &&

#解锁根用户
./unlock_root.sh &&

#部署系统合约
./deploy-system-contracts.sh &&

# cd nodeManager &&
#在节点管理合约中添加主节点和从节点
bash  ./add-all-nodes-init.sh 
bash  ./add-all-nodes-distrib.sh
bash  ./start_all.sh

sleep 10
#将从节点改为共识节点
./update-all.sh


