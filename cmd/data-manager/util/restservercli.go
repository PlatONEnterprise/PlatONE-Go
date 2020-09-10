package util

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

type cnsInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Address    string `json:"address"`
	Owner      string `json:"origin"`
	CreateTime int64  `json:"create_time"`
}

func GetAllCNS() ([]*cnsInfo, error) {
	//todo

	var cnses []*cnsInfo
	return cnses, nil
}

func GetLatestCNS(name string) (*cnsInfo, error) {
	//todo

	var cns cnsInfo
	return &cns, nil
}

func GetCNSByAddress(addr string) (*cnsInfo, error) {
	var cns cnsInfo
	return &cns, nil
}
