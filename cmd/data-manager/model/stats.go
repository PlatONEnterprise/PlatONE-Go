package model

import (
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
)

func GetStats(c *dbCtx.Context) (*Stats, error) {
	var (
		stats Stats
		err   error
	)

	stats.LatestBlock, err = DefaultBlock.LatestHeight(c)
	if nil != err {
		logrus.Errorln(err)
		return nil, err
	}

	stats.TotalContract, err = DefaultTx.TotalContract(c)
	if nil != err {
		logrus.Errorln(err)
		return nil, err
	}

	stats.TotalTx, err = DefaultTx.TotalTx(c)
	if nil != err {
		logrus.Errorln(err)
		return nil, err
	}

	stats.TotalNode, err = DefaultNode.AllNodeCount(c)
	if nil != err {
		logrus.Errorln("failed to find amount of nodes,err:", err)
		return nil, err
	}

	return &stats, nil
}

//
//import (
//	"context"
//	dbCtx "data-manager/db/context"
//	"github.com/sirupsen/logrus"
//	"go.mongodb.org/mongo-driver/bson"
//	"go.mongodb.org/mongo-driver/bson/primitive"
//	"go.mongodb.org/mongo-driver/mongo/options"
//	"time"
//)
//
//const (
//	collectionNameStats = "stats"
//)
//
//var (
//	StatsDBID = primitive.NewObjectID().Hex()
//)
//
//var DefaultStats = newStats()
//
//type stats struct{}
//
//func newStats() *stats {
//	return new(stats)
//}
//
//func (this *stats) Stats(c *dbCtx.Context) (*Stats, error) {
//	collection := c.Collection(collectionNameStats)
//	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
//
//	var s Stats
//	err := collection.FindOne(
//		ctx,
//		bson.D{{}}).Decode(&s)
//	if nil != err {
//		logrus.Errorln(err)
//		return nil, err
//	}
//
//	logrus.Debugf("Stats:%+v", s)
//	return &s, nil
//}
//
//func (this *stats) UpsertStats(c *dbCtx.Context, stats *Stats) error {
//	collection := c.Collection(collectionNameStats)
//	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
//
//	update := bson.M{"$set": stats}
//	updateOpts := options.Update().SetUpsert(true)
//
//	_, err := collection.UpdateOne(ctx, bson.M{"_id": StatsDBID}, update, updateOpts)
//	if nil != err {
//		logrus.Errorln(err)
//		return err
//	}
//
//	return nil
//}
