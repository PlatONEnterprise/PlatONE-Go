.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    Bcos-go/cmd/swarm
	go build -o bin/geth     Bcos-go/cmd/geth
	go build -o bin/bootnode Bcos-go/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v Bcos-go/swarm/...
