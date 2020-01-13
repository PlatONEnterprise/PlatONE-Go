# Changelog

## [unknown]
### Features
* [contract] Wasm合约支持float型计算 -- 王琪，王忠莉，朱冰心，潘晨，吴启迪，杜满想

### Bug Fixes
* [ctool] 返回值是uint32类型时无法解析。 -- 杜满想
* [contract] Wasm合约无法打印uint64类型的变量。 -- 王琪，杜满想

## [0.9.1] -- 2019-11-22
### Breaking Changes
### Features
* [contract] 调用一个没有在CNS中注册的合约时报错，receipt的status设为false。-- 简海波，葛鑫

### Improvements
* [contract] 简化了wasm与solidity兼容调用方式。-- 汤勇
* [chain] 添加对版本的支持，`./platonce  --version`可以打印当前版本。 -- 葛鑫
* [contract] 删除sm密码库的静态库文件，改为用源码编译，方便为以后的跨平台做准备。 -- 潘晨
* [chain] 暂时注释掉VC（verifiable computation，可验证计算）和nizkpail相关代码为跨平台做准备。-- 杜满想

### Bug Fixes
* [node] 以前在节点管理合约中, 删除节点后, 此节点就无法恢复正常状态.  目前 是支持可以更改为正常状态了。 -- 汤勇
* [chain] 各个节点时间不一致情况下搭链，搭链成功后向cnsManager合约注册一个新合约，会导致各节点状态不一致。  -- 葛鑫
* [contract] 合约中Event参数不支持int类型。 -- 葛鑫
* [contract] 合约中调用assert方法无法打印。-- 黄赛杰