package syncer

import (
	"data-manager/config"
	"data-manager/db"
	dbCtx "data-manager/db/context"
	"data-manager/model"
	"data-manager/util"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

var DefaultSyncer = newSyncer()

type syncer struct {
	stop  chan int
	dbCtx *dbCtx.Context
}

func newSyncer() *syncer {
	return &syncer{
		stop:  make(chan int),
		dbCtx: dbCtx.New(db.DefaultDB),
	}
}

func (this *syncer) Run() {
	go this.loop()
	logrus.Info("start to sync.")
}

func (this *syncer) loop() {
	tick := time.NewTicker(config.Config.SyncConf.SyncInterval())
	target := time.Until(config.Config.SyncTxCountConf.GetWhen())
	syncTxCountTick := time.NewTimer(target)

	for {
		select {
		case <-tick.C:
			this.sync()

		case <-syncTxCountTick.C:
			for i := 0; i < config.Config.SyncTxCountConf.TryTimes; i++ {

				err := this.syncTxStats()
				if nil != err {
					logrus.Errorln(err)
					continue
				}

				break
			}

			syncTxCountTick.Reset(time.Until(config.Config.SyncTxCountConf.GetWhen().AddDate(0, 0, 1)))

		case <-this.stop:
			logrus.Info("sync stop")
			syncTxCountTick.Stop()
			tick.Stop()

			return
		}
	}
}

func (this *syncer) sync() {
	err := this.syncNodes()
	if nil != err {
		logrus.Errorln("failed to sync nodes,err:", err)
		//return
	} else {
		logrus.Debug("sync nodes success.")
	}

	err = this.syncCNS()
	if nil != err {
		logrus.Errorln("failed to sync cns,err:", err)
		//return
	} else {
		logrus.Debug("sync cns success.")
	}

	err = this.syncBlocks()
	if nil != err {
		logrus.Errorln("failed to sync blocks,err:", err)
		return
	}
	logrus.Debug("sync blocks success.")

	//err = this.syncStats()
	//if nil != err {
	//	logrus.Errorln("failed to sync stats,err:", err)
	//	return
	//}
}

func (this *syncer) syncBlocks() error {
	block, err := util.DefaultNode.LatestBlock()
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	dbBlock, err := model.DefaultBlock.LatestBlock(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return this.doSyncBlocks(block.NumberU64(), dbBlock.Height)
}

func (this *syncer) doSyncBlocks(heightTarget, heightCur uint64) error {
	for i := heightCur + 1; i <= heightTarget; i++ {
		block, err := util.DefaultNode.BlockByHeight(i)
		if nil != err {
			logrus.Errorln(err)
			return err
		}

		err = this.doSyncTxs(block)
		if nil != err {
			logrus.Errorln(err)
			return err
		}

		var dbBlock model.Block

		dbBlock.Height = i
		dbBlock.ExtraData = hex.EncodeToString(block.Extra())
		dbBlock.GasLimit = block.GasLimit()
		dbBlock.GasUsed = block.GasUsed()
		dbBlock.Hash = block.Hash().Hex()
		dbBlock.ParentHash = block.ParentHash().Hex()
		dbBlock.Proposer = block.Coinbase().Hex()
		dbBlock.Timestamp = block.Time().Int64()
		dbBlock.TxAmount = uint64(block.Transactions().Len())
		dbBlock.Size = block.Size().String()

		err = model.DefaultBlock.InsertBlock(this.dbCtx, &dbBlock)
		if nil != err {
			logrus.Errorln(err)
			return err
		}
	}

	return nil
}

func (this *syncer) doSyncTxs(block *types.Block) error {
	for _, tx := range block.Transactions() {
		var dbTx model.Tx

		dbTx.Timestamp = block.Time().Int64()
		dbTx.Hash = tx.Hash().Hex()
		dbTx.GasLimit = tx.Gas()
		receipt, err := util.DefaultNode.TransactionReceipt(tx.Hash())
		if nil != err {
			logrus.Errorln("fail to get transaction receipt.err:", err)
			return err
		}
		var recpt model.Receipt
		recpt.GasUsed = receipt.GasUsed
		recpt.ContractAddress = receipt.ContractAddress.String()
		if common.IsHexZeroAddress(recpt.ContractAddress) {
			recpt.ContractAddress = ""
		}

		bin, err := json.Marshal(receipt.Logs)
		if nil != err {
			logrus.Errorln(err)
			return err
		}
		recpt.Event = string(bin) //todo parse event
		recpt.Status = receipt.Status

		dbTx.Receipt = &recpt
		from, err := util.Sender(tx)
		if nil != err {
			logrus.Errorln("fail to get sender of tx.err:", err)

			return err
		}
		dbTx.From = from.Hex()

		dbTx.GasPrice = tx.GasPrice().Uint64()
		dbTx.Input = hex.EncodeToString(tx.Data())
		dbTx.Nonce = fmt.Sprintf("%d", tx.Nonce())
		dbTx.To = ""
		if tx.To() != nil {
			dbTx.To = tx.To().Hex()
		}
		dbTx.Typ = tx.Type()
		dbTx.Value = tx.Value().Uint64()
		dbTx.Height = block.NumberU64()

		err = model.DefaultTx.InsertTx(this.dbCtx, &dbTx)
		if nil != err {
			logrus.Errorln(err)
			return err
		}
	}

	return nil
}

func (this *syncer) syncTxStats() error {
	now := time.Now().AddDate(0, 0, -1)
	y, m, d := now.Date()

	start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	end := time.Date(y, m, d, 23, 59, 59, 0, time.Local)
	amount, err := model.DefaultTx.TxAmountByTime(this.dbCtx, start.Unix(), end.Unix())
	if nil != err {
		return err
	}

	err = model.DefaultTxStats.UpsertTxAmountOneDay(this.dbCtx, fmt.Sprintf("%d:%d:%d", y, m, d), amount)
	if nil != err {
		return err
	}

	return nil
}

//
//func (this *syncer) syncStats() error {
//	var stats model.Stats
//	var err error
//	stats.LatestBlock, err = model.DefaultBlock.LatestHeight(this.dbCtx)
//	if nil != err {
//		logrus.Errorln(err)
//		return err
//	}
//
//	stats.TotalContract, err = model.DefaultTx.TotalContract(this.dbCtx)
//	if nil != err {
//		logrus.Errorln(err)
//		return err
//	}
//
//	stats.TotalTx, err = model.DefaultTx.TotalTx(this.dbCtx)
//	if nil != err {
//		logrus.Errorln(err)
//		return err
//	}
//
//	stats.TotalNode, err = model.DefaultNode.AllNodeCount(this.dbCtx)
//	if nil != err {
//		logrus.Errorln("failed to find amount of nodes,err:", err)
//		return err
//	}
//
//	err = model.DefaultStats.UpsertStats(this.dbCtx, &stats)
//	if nil != err {
//		logrus.Errorln(err)
//		return err
//	}
//
//	return nil
//}

func (this *syncer) syncNodes() error {
	nodeInfos, err := util.GetNodes()
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	nodes := make([]*model.Node, 0, len(nodeInfos))
	for _, info := range nodeInfos {
		var node model.Node
		node.Typ = info.Typ
		node.Name = info.Name
		node.P2PPort = info.P2PPort
		node.InternalIP = info.InternalIP
		node.Desc = info.Desc
		node.ExternalIP = info.ExternalIP
		node.PubKey = info.PubKey
		node.RPCPort = info.RPCPort
		node.Typ = info.Typ
		node.Status = info.Status
		node.Owner = info.Owner

		node.IsAlive = util.IsNodeAlive(info)

		nodes = append(nodes, &node)
	}

	//TODO better idea
	err = model.DefaultNode.DeleteAllNodes(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return model.DefaultNode.InsertNodes(this.dbCtx, nodes)
}

func (this *syncer) syncCNS() error {
	cnses, err := util.GetAllCNS()
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	mapCns := map[string]*model.CNS{}
	for _, info := range cnses {
		latest, err := util.GetLatestCNS(info.Name)
		if nil != err {
			logrus.Errorln(err)
			return err
		}

		cns, ok := mapCns[info.Name]
		if !ok {
			cns = &model.CNS{Infos: []*model.CNSInfo{}}
			mapCns[info.Name] = cns
		}

		if info.Name == latest.Name && strings.ToLower(info.Address) == strings.ToLower(latest.Address) {
			cns.Address = info.Address
			cns.Name = info.Name
			cns.Version = info.Version
		}

		var ci model.CNSInfo
		ci.Version = info.Version
		ci.Address = info.Address

		cns.Infos = append(cns.Infos, &ci)
	}

	var modelCnses []*model.CNS
	for _, v := range mapCns {
		modelCnses = append(modelCnses, v)
	}

	//TODO better idea
	err = model.DefaultCNS.DeleteAllCNS(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	sort.SliceStable(modelCnses, func(i, j int) bool {
		return strings.Compare(modelCnses[i].Name, modelCnses[j].Name) > 0
	})

	return model.DefaultCNS.InsertCNS(this.dbCtx, modelCnses)
}
