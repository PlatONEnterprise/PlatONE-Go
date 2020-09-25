package model

import (
	"context"
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const (
	collectionNameCNS = "cns"
)

var DefaultCNS = newCNS()

type cns struct{}

func newCNS() *cns {
	return new(cns)
}

func (this *cns) InsertCNS(c *dbCtx.Context, infos []*CNS) error {
	if len(infos) == 0 {
		return nil
	}

	collection := c.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	docs := []interface{}{}
	for _, info := range infos {
		docs = append(docs, info)
	}
	_, err := collection.InsertMany(ctx, docs)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *cns) DeleteAllCNS(c *dbCtx.Context) error {
	collection := c.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	//delete all docs
	_, err := collection.DeleteMany(ctx, bson.D{{}})
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *cns) QueryCNS(c *dbCtx.Context, pageIndex, pageSize int64) ([]*CNS, error) {
	collection := c.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.D{{}}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bson.D{{"name", -1}})

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := []*CNS{}
	err = cur.All(ctx, &results)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return results, nil
}

func (this *cns) Total(c *dbCtx.Context) (int64, error) {
	collection := c.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.M{}
	amount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logrus.Errorln(err)
		return 0, err
	}

	return amount, nil
}
