# Changelog

## [unknow] -- unknow

### Breaking Changes

### Features

### Improvements

### Bug Fixes

* [chain] 修复第一个节点数据清空无法再加入网络的问题。--汤勇

## [0.9.4] -- 2019-12-20
### Breaking Changes

### Features

### Improvements

### Bug Fixes

* [contract] 修复合约调用ecrecover时，若签名无效，则虚拟机执行失败的问题，现改为返回nil。--潘晨

## [0.9.3] -- 2019-12-11

### Breaking Changes

### Features

### Improvements

### Bug Fixes

* [chain] 并发访问所有链接的节点map集合时，出现并发读写错误，导致节点宕机。--汤勇

## [0.9.2] -- 2019-12-06

### Breaking Changes
### Features
### Improvements
### Bug Fixes
* [chain] 区块执行时间过长时，共识无法正常工作，不能继续出块。--葛鑫

## [0.9.1] -- 2019-11-22
### Breaking Changes
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