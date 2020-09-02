package syncer

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type nodeInfo struct {
	Name   string `json:"name"`
	PubKey string `json:"pub_key"`
	Desc   string `json:"desc"`
	//IsAlive    bool   `json:"is_alive"`
	InternalIP string `json:"internal_ip"`
	ExternalIP string `json:"external_ip"`
	RPCPort    int    `json:"rpc_port"`
	P2PPort    int    `json:"p2p_port"`
	Typ        int    `json:"type"`
}

func GetAmountOfNodes() (int, error) {
	nodes, err := GetNodes()
	if nil != err {
		logrus.Errorln(err)
		return 0, err
	}

	return len(nodes), nil
}

func GetNodes() ([]*nodeInfo, error) {
	//TODO
	var nodes []*nodeInfo
	return nodes, nil
}

func IsNodeAlive(info *nodeInfo) bool {
	address := fmt.Sprintf("%s:%d", info.InternalIP, info.P2PPort)
	conn, err := net.Dial("tcp", address)
	if nil != err {
		return false
	}
	defer conn.Close()

	timeout := time.Second * 5
	err = conn.SetWriteDeadline(time.Now().Add(timeout))
	if nil != err {
		return false
	}

	_, err = conn.Write([]byte("ping"))
	if nil != err {
		return false
	}

	return true
}
