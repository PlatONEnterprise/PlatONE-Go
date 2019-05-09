.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    PlatONE-go/cmd/swarm
	go build -o bin/geth     PlatONE-go/cmd/geth
	go build -o bin/bootnode PlatONE-go/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v PlatONE-go/swarm/...
