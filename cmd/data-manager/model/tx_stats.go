package model

import (
	"context"
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	collectionNameTxStats = "tx_stats"
)

var DefaultTxStats = newTxStats()

type txStats struct {
}

func newTxStats() *txStats {
	return new(txStats)
}

func (this *txStats) UpsertTxAmountOneDay(c *dbCtx.Context, date string, amount int64) error {
	collection := c.Collection(collectionNameTxStats)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	stats := TxStats{}
	stats.Date = date
	stats.TxAmount = amount

	update := bson.M{"$set": stats}
	updateOpts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, bson.M{"date": date}, update, updateOpts)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *txStats) History(c *dbCtx.Context, num int64) ([]*TxStats, error) {
	collection := c.Collection(collectionNameTxStats)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.D{{}}
	findOps := buildOptionsByQuery(1, num)
	findOps.SetSort(bson.D{{"date", -1}})

	cur, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := []*TxStats{}
	err = cur.All(ctx, &results)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return results, nil
}
