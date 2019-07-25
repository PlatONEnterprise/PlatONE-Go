#!/bin/bash


echo '
###########################################
####        build 4-node chain         ####
###########################################
'

./setup-genesis.sh --ip 127.0.0.1 --p2p_port 16791 --auto "true"

./init-node.sh --nodeid 0 --ip 127.0.0.1 --rpc_port 6791 --p2p_port 16791 --ws_port 26791 --auto "true"
./init-node.sh --nodeid 1 --ip 127.0.0.1 --rpc_port 6792 --p2p_port 16792 --ws_port 26792 --auto "true"
./init-node.sh --nodeid 2 --ip 127.0.0.1 --rpc_port 6793 --p2p_port 16793 --ws_port 26793 --auto "true"
./init-node.sh --nodeid 3 --ip 127.0.0.1 --rpc_port 6794 --p2p_port 16794 --ws_port 26794 --auto "true"

./start-node.sh --nodeid 0 

./deploy-system-contract.sh --auto "true"

./start-node.sh --nodeid 1
./start-node.sh --nodeid 2
./start-node.sh --nodeid 3

./add-node.sh --nodeid 1
./add-node.sh --nodeid 2
./add-node.sh --nodeid 3


sleep 10

./update_to_consensus_node.sh --nodeid 0
./update_to_consensus_node.sh --nodeid 1
./update_to_consensus_node.sh --nodeid 2
./update_to_consensus_node.sh --nodeid 3
