package syncer

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/config"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/db"
	dbCtx "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/db/context"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/model"
	"encoding/hex"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/sirupsen/logrus"
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

func (this *syncer) exec() {
	err := this.syncNodes()
	if nil != err {
		logrus.Errorln(err)
		//return
	}

	err = this.syncBlocks()
	if nil != err {
		logrus.Errorln(err)
		return
	}

	this.syncStats()
}

func (this *syncer) syncBlocks() error {
	block, err := defaultNode.LatestBlock()
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
	for i := heightCur + 1; i < heightTarget; i++ {
		block, err := defaultNode.BlockByHeight(i)
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
		dbBlock.Proposer = block.Coinbase().Hex() //TODO
		dbBlock.Timestamp = block.Time().Int64()
		dbBlock.TxAmount = uint64(block.Transactions().Len())

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
		receipt, err := defaultNode.TransactionReceipt(tx.Hash())
		if nil != err {
			logrus.Errorln(err)
			return err
		}
		var recpt model.Receipt
		recpt.GasUsed = receipt.GasUsed
		recpt.ContractAddress = receipt.ContractAddress.Hex()
		//recpt.Event = receipt.Logs[0].Address //TODO
		recpt.Status = receipt.Status

		dbTx.Receipt = &recpt
		from, err := types.FrontierSigner{}.Sender(tx)
		if nil != err {
			logrus.Errorln(err)

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

		err = model.DefaultTx.InsertTx(this.dbCtx, &dbTx)
		if nil != err {
			logrus.Errorln(err)
			return err
		}
	}

	return nil
}

func (this *syncer) syncStats() error {
	var stats model.Stats
	var err error
	stats.LatestBlock, err = model.DefaultBlock.LatestHeight(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	stats.TotalContract, err = model.DefaultTx.TotalContract(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	stats.TotalTx, err = model.DefaultTx.TotalTx(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	totalNode, err := GetAmountOfNodes()
	if nil != err {
		logrus.Errorln(err)
		return err
	}
	stats.TotalNode = uint64(totalNode)

	err = model.DefaultStats.UpsertStats(this.dbCtx, &stats)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *syncer) syncNodes() error {
	//TODO better idea
	err := model.DefaultNode.DeleteAllNodes(this.dbCtx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	nodeInfos, err := GetNodes()
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	nodes := make([]*model.Node, 0, len(nodeInfos))
	for _, info := range nodeInfos {
		//TODO
		var node model.Node
		node.Typ = info.Typ

		node.IsAlive = IsNodeAlive(info)
	}

	return model.DefaultNode.InsertNodes(this.dbCtx, nodes)
}

func (this *syncer) loop() {
	tick := time.NewTicker(config.Config.SyncConf.SyncInterval())

	for {
		select {
		case <-tick.C:
			this.exec()

		case <-this.stop:
			logrus.Info("sync stop")
			return
		}
	}
}
