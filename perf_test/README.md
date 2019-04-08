# 测试简单合约

1. 部署简单合约
```shell
# ctool deploy --abi ./data/demo.cpp.abi.json --code ./data/demo.wasm --config ./data/config.json

trasaction hash: 0xab39916045dc1c64af95acb69927400514899509b5ab8b874549e7fab5064b7b
contract address: 0x2124e0d7392683a9fac7167e30da82858bd0f514
```

2. 调用perf_test进行性能测试（以下命令均执行于当前文档所在目录）
方法一实时（推荐）：
```shell
# ./perf_test -stressTest=1 -abiPath="./data/demo.cpp.abi.json" -configPath="./data/config.json"  -contractAddress="0x2124e0d7392683a9fac7167e30da82858bd0f514" -registerContractNum=2000 -realtimeTps=true -consensusTest=true
```

方法二非实时：
```shell
# ./perf_test -stressTest=1 -abiPath="./data/demo.cpp.abi.json" -configPath="./data/config.json"  -contractAddress="0x2124e0d7392683a9fac7167e30da82858bd0f514" -registerContractNum=2000
```


# 测试复杂合约

1. 部署复杂合约
```shell
# ctool deploy --abi ./data/nodeRegister.cpp.abi.json --code ./data/nodeRegister.wasm --config ./data/config.json

trasaction hash: 0x83d5c1ba5c76bd91efdbdf34f549277314d71fe0587a21c0b8f8d02c064785e5
contract address: 0x7cf06df7bcb5291007ff04f69c179e07a2e1b640
```

2. 调用perf_test进行性能测试
方法一实时（推荐）：
```shell
# ./perf_test -stressTest=2 -abiPath="./data/cnsManager.cpp.abi.json" -configPath="./data/config.json" -registerContractNum=100 -deployContractAddress="0x7cf06df7bcb5291007ff04f69c179e07a2e1b640" -realtimeTps=true -consensusTest=true
```

方法二非实时：
```shell
# ./perf_test -stressTest=2 -abiPath="./data/cnsManager.cpp.abi.json" -configPath="./data/config.json" -registerContractNum=100 -deployContractAddress="0x7cf06df7bcb5291007ff04f69c179e07a2e1b640"
```

# 注意： data目录下config.json文件里请填写自己的账户信息与节点url，并事先在相应节点解锁。