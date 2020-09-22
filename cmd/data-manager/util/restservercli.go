package util

import (
	"data-manager/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type nodeInfo struct {
	Name   string `json:"name"`
	PubKey string `json:"publicKey"`
	Desc   string `json:"desc"`
	//IsAlive    bool   `json:"is_alive"`
	InternalIP string `json:"internalIP"`
	ExternalIP string `json:"externalIP"`
	RPCPort    int    `json:"rpcPort"`
	P2PPort    int    `json:"p2pPort"`
	Typ        int    `json:"type"`
	Status     int    `json:"status"`
	Owner      string `json:"owner"`
}

type nodeResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []*nodeInfo `json:"data"`
}

func GetAmountOfNodes() (int, error) {
	nodes, err := GetNodes()
	if nil != err {
		logrus.Errorln("failed to get nodes,err:", err)
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
		"/cns/mappings",
		name,
		config.Config.ChainConf.NodeRpcAddress,
	)

	var ret string
	err := httpGet(url, &ret)
	if nil != err {
		return nil, err
	}

	var ci cnsInfo
	ci.Name = name
	ci.Address = ret

	return &ci, nil
}

var (
	ErrCNSNotFound = errors.New("cns not found")
)

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

	if 0 == len(ret) {
		return nil, ErrCNSNotFound
	}

	return ret[0], nil
}

func urlCnsComponents(url string) ([]*cnsInfo, error) {
	var ret cnsResult

	err := httpGet(url, &ret)
	if nil != err {
		return nil, err
	}

	if ret.Code == 1 {
		logrus.Warningln("cns not found,url:", url, " result:", ret)
		return []*cnsInfo{}, nil
	} else if ret.Code != 0 {
		err := errors.New(ret.Msg)
		logrus.Errorln("get cns, result:", ret, "err:", err)
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
	logrus.Debugln("url:", url, "httpget body:", string(bin))

	err = json.Unmarshal(bin, ret)
	if err != nil {
		logrus.Errorln("failed to unmarshal data that from resp.body,err:", err, "resp data:", string(bin))
		return err
	}

	return nil
}
