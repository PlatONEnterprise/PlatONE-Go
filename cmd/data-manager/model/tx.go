package model

import (
	"context"
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

const (
	collectionNameTxs = "txs"
)

var DefaultTx = newTx()

type tx struct {
}

func newTx() *tx {
	return new(tx)
}

func (this *tx) InsertTx(c *dbCtx.Context, tx *Tx) error {
	collection := c.Collection(collectionNameTxs)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	_, err := collection.InsertOne(ctx, tx)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *tx) TotalTx(c *dbCtx.Context) (int64, error) {
	filter := bson.D{}

	return this.totalTxByFilter(c, filter)
}

func (this *tx) TotalContract(c *dbCtx.Context) (int64, error) {
	filter := bson.M{}
	filter["receipt.contract_address"] = bson.M{"$ne": ""}

	return this.totalTxByFilter(c, &filter)
}

func (this *tx) totalTxByFilter(c *dbCtx.Context, filter interface{}) (int64, error) {
	collection := c.Collection(collectionNameTxs)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	count, err := collection.CountDocuments(ctx, filter)
	if nil != err {
		logrus.Errorln(err)
		return 0, err
	}

	return count, nil
}

func (this *tx) TxsInHeight(c *dbCtx.Context, pageIndex, pageSize int64, blockHeight uint64) ([]*Tx, error) {
	filter := bson.M{"block_height": blockHeight}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bsonx.Doc{{"timestamp", bsonx.Int32(-1)}})

	return this.txs(c, filter, findOps)
}

func (this *tx) Contracts(c *dbCtx.Context, pageIndex, pageSize int64) ([]*Tx, error) {
	filter := bson.M{}
	filter["receipt.contract_address"] = bson.M{"$ne": ""}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bsonx.Doc{{"timestamp", bsonx.Int32(-1)}})

	return this.txs(c, filter, findOps)
}

func (this *tx) TxsFromAddress(c *dbCtx.Context, pageIndex, pageSize int64, addres string) ([]*Tx, error) {
	filter := bson.M{"from": addres}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bsonx.Doc{{"timestamp", bsonx.Int32(-1)}})

	return this.txs(c, filter, findOps)
}

func (this *tx) Txs(c *dbCtx.Context, pageIndex, pageSize int64) ([]*Tx, error) {
	filter := bson.D{{}}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bson.D{{"timestamp", -1}})

	return this.txs(c, filter, findOps)
}

func (this *tx) txs(c *dbCtx.Context, filter interface{}, findOps *options.FindOptions) ([]*Tx, error) {
	collection := c.Collection(collectionNameTxs)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	cur, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := []*Tx{}
	err = cur.All(ctx, &results)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return results, nil
}

func (this *tx) TxAmountByTime(c *dbCtx.Context, start, end int64) (int64, error) {
	collection := c.Collection(collectionNameTxs)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.M{}
	filter["timestamp"] = bson.M{"$gte": start, "$lte": end}
	amount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logrus.Errorln(err)
		return 0, err
	}

	return amount, nil
}

func (this *tx) ContractByAddress(c *dbCtx.Context, addr string) (*Tx, error) {
	filter := bson.M{}
	filter["receipt.contract_address"] = addr

	return this.queryTx(c, filter)
}

func (this *tx) TxByHash(c *dbCtx.Context, hash string) (*Tx, error) {
	filter := bson.M{}
	filter["tx_hash"] = hash

	return this.queryTx(c, filter)
}

func (this *tx) queryTx(c *dbCtx.Context, filter bson.M) (*Tx, error) {
	collection := c.Collection(collectionNameTxs)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var t Tx
	err := collection.FindOne(ctx, filter).Decode(&t)
	if nil != err {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("Tx:%+v", t)
	return &t, nil
}
