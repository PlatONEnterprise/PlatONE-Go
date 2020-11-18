# Changelog

## [unknown]

### Breaking Changes
* [system contract] 系统合约重构成预编译合约形式
* [other] 删除eip，DAO等版本升级的Hard Fork和兼容性检查;
* [other] 删除Rinkeby，Testnet;删除ChainConfig的EmptyBlock设置;删除Clique；删除difficulty；删除dev模式；
* [other] 删除默认配置,并重写了genesis初始化逻辑。

### Improvements
* [other] 版本管理采用mod模型 - 汤涌，于宗坤，杜满想
* [other] 删除whisper,swarm,mobile,cmd/wnode（Whisper node）- 杜满想
* [other] 删除pow相关逻辑(reorg,sidechain),删除cbft - 杜满想

### Features
* [chain] 添加一链多账本功能（群组预编译化系统合约等等）
* [other] 可视化运维平台


## [0.9.12] 2020-08-25
### Breaking Changes
### Improvements
* [chain] genesis时间戳自动设置为当前系统时间 --葛鑫

### Features
### Bug Fixes

## [0.9.11] 2020-06-04
### Breaking Changes
### Improvements
### Features
* [chain] WASM虚拟机对大浮点数和大整数的支持 --于宗坤
* [other] PlatONE-CDT

### Bug Fixes


## [0.9.10] 2020-05-07
### Breaking Changes
### Improvements
### Features
* [chain] 交易处理執行時gasPrice寫入receipt功能。 --潘晨

### Bug Fixes
* [chain] import功能bug修复。  --葛鑫
* [chain] 解决PlatONE终端log输出时，日志等级设置失效问题。 --汤勇

## [0.9.9]
### Breaking Changes
### Improvements
* [chain] 在共识模块中直接同步写入区块，以提高区块链交易处理性能。 --葛鑫
### Features
* [chain] 交易处理流程引入根据交易消耗gas扣除用户特定token的功能。 --潘晨
### Bug Fixes

## [0.9.8]
### Breaking Changes
### Features
* [contract] Wasm合约支持float型计算 -- 王琪，王忠莉，朱冰心，潘晨，吴启迪，杜满想

### Bug Fixes
* [ctool] 返回值是uint32类型时无法解析。 -- 杜满想
* [contract] Wasm合约无法打印uint64类型的变量。 -- 王琪，杜满想
* [contract]　修复sm2验签某些公钥解析失败的bug。--潘晨

## [0.9.7] -- 2020-01-16
### Bug Fixes
* [contract]　修改secp256k1验签功能所用的hash函數。--潘晨

## [0.9.6] -- 2020-01-15
### Features
* [contract]　增加secp256k1和r1的验签功能。--潘晨

### Bug Fixes
* [chain] txpool先去重再验签。 -- 葛鑫

## [0.9.5] -- 2020-01-03
### Bug Fixes
* [chain] 修复第一个节点数据清空无法再加入网络的问题。--汤勇
* [chain] 共识模块中共识结束直接写入区块数据，可能会造成并发问题，修改为由p2p.fetcher异步写入。 --葛鑫 
* [chain] 共识消息处理中，投票类消息单独用一个Event Channel消息处理。 -- 葛鑫

## [0.9.4] -- 2020-12-20
### Bug Fixes
* [contract] 修复合约调用ecrecover时，若签名无效，则虚拟机执行失败的问题，现改为返回nil。--潘晨

## [0.9.3] -- 2019-12-11
### Bug Fixes
* [chain] 并发访问所有链接的节点map集合时，出现并发读写错误，导致节点宕机。--汤勇

## [0.9.2] -- 2019-12-06
### Bug Fixes
* [chain] 区块执行时间过长时，共识无法正常工作，不能继续出块。--葛鑫

## [0.9.1] -- 2019-11-22
### Features
* [contract] 调用一个没有在CNS中注册的合约时报错，receipt的status设为false。-- 简海波，葛鑫

### Improvements
* [contract] 简化了wasm与solidity兼容调用方式。-- 汤勇
* [chain] 添加对版本的支持，`./platone  --version`可以打印当前版本。 -- 葛鑫
* [contract] 删除sm密码库的静态库文件，改为用源码编译，方便为以后的跨平台做准备。 -- 潘晨
* [chain] 暂时注释掉VC（verifiable computation，可验证计算）和nizkpail相关代码为跨平台做准备。-- 杜满想

### Bug Fixes
* [node] 以前在节点管理合约中, 删除节点后, 此节点就无法恢复正常状态.  目前 是支持可以更改为正常状态了。 -- 汤勇
* [chain] 各个节点时间不一致情况下搭链，搭链成功后向cnsManager合约注册一个新合约，会导致各节点状态不一致。  -- 葛鑫
* [contract] 合约中Event参数不支持int类型。 -- 葛鑫
* [contract] 合约中调用assert方法无法打印。-- 黄赛杰
