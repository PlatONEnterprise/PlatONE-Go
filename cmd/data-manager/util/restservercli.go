package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/config"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
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

type nodeResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []*nodeInfo `json:"data"`
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
	url := fmt.Sprintf(
		"%s%s?endpoint=%s",
		config.Config.ChainConf.NodeRestServer,
		"/node/components",
		config.Config.ChainConf.NodeRpcAddress,
	)

	return urlNodeComponents(url)
}

func urlNodeComponents(url string) ([]*nodeInfo, error) {
	var ret nodeResult

	err := httpGet(url, &ret)
	if nil != err {
		return nil, err
	}

	if ret.Code != 0 {
		err := errors.New("node not found,msg:" + ret.Msg)
		logrus.Errorln(err)
		return nil, err
	}

	return ret.Data, nil
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

type cnsResult struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []*cnsInfo `json:"data"`
}

func GetAllCNS() ([]*cnsInfo, error) {
	url := fmt.Sprintf(
		"%s%s?endpoint=%s",
		config.Config.ChainConf.NodeRestServer,
		"/cns/components",
		config.Config.ChainConf.NodeRpcAddress,
	)

	ret, err := urlCnsComponents(url)
	if nil != err {
		return nil, err
	}

	return ret, nil
}

func GetLatestCNS(name string) (*cnsInfo, error) {
	url := fmt.Sprintf(
		"%s%s/%s?endpoint=%s&version=latest",
		config.Config.ChainConf.NodeRestServer,
		"/cns/mapings",
		name,
		config.Config.ChainConf.NodeRpcAddress,
	)

	ret := struct {
		Result string `json:"cnsResult"`
	}{}

	err := httpGet(url, &ret)
	if nil != err {
		return nil, err
	}

	var ci cnsInfo
	ci.Name = name
	ci.Address = ret.Result

	return &ci, nil
}

func GetCNSByAddress(addr string) (*cnsInfo, error) {
	url := fmt.Sprintf(
		"%s%s?endpoint=%s&address=%s",
		config.Config.ChainConf.NodeRestServer,
		"/cns/components",
		config.Config.ChainConf.NodeRpcAddress,
		addr,
	)

	ret, err := urlCnsComponents(url)
	if nil != err {
		return nil, err
	}

	return ret[0], nil
}

func urlCnsComponents(url string) ([]*cnsInfo, error) {
	var ret cnsResult

	err := httpGet(url, &ret)
	if nil != err {
		return nil, err
	}

	if ret.Code != 0 {
		err := errors.New("cns not found,msg:" + ret.Msg)
		logrus.Errorln(err)
		return nil, err
	}

	return ret.Data, nil
}

func httpGet(url string, ret interface{}) error {
	resp, err := http.Get(url)
	if nil != err {
		logrus.Errorln("failed to http get,err:", err)
		return err
	}
	defer resp.Body.Close()

	bin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("failed to read from resp.body,err:", err)
		return err
	}

	err = json.Unmarshal(bin, ret)
	if err != nil {
		logrus.Errorln("failed to unmarshal data that from resp.body,err:", err, "resp data:", string(bin))
		return err
	}

	return nil
}
