## bcWasm System Contract

Welcome to the System Contract source code repository! It's based on BCWasm.

## Building the source
* Building system contracts for the first time requires you to do the following steps:
```shell
git clone https://180.167.100.189:20443/PlatONE/src/node/SysContracts.git
rm -rf ./build
./script/autoproject.sh .
```

* Later if you want to specified project, you can run this cmd in ./build:
```shell
make cnsManager
```

* Create New SystemContract and build
``` shell
./script/autoproject.sh . newSystemContractName
```
This Cmd will create new dir under ./systemContract and ./build/systemContract.
Now the contract `./systemContract/newSystemContractName/newSystemContractName.cpp` is empty, you should write contract code into this file, and then go to `./build` dir then `make newSystemContractName`

* testContract
`./script/autoprojectForTest.sh` can be used to create contract for testing in testContract. 