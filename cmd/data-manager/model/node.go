package model

import (
	"context"
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const (
	collectionNameNodes = "nodes"
)

var DefaultNode = newNode()

type node struct{}

func newNode() *node {
	return new(node)
}

func (this *node) InsertNodes(c *dbCtx.Context, nodes []*Node) error {
	collection := c.Collection(collectionNameNodes)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	docs := []interface{}{}
	for _, node := range nodes {
		docs = append(docs, node)
	}
	_, err := collection.InsertMany(ctx, docs)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *node) DeleteAllNodes(c *dbCtx.Context) error {
	collection := c.Collection(collectionNameNodes)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	//delete all docs
	_, err := collection.DeleteMany(ctx, bson.D{{}})
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *node) AllNodeCount(c *dbCtx.Context) (int, error) {
	nodes, err := this.AllNodes(c)
	if nil != err {
		return 0, err
	}

	return len(nodes), nil
}

func (this *node) AllNodes(c *dbCtx.Context) ([]*Node, error) {
	collection := c.Collection(collectionNameNodes)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.D{{}}

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := []*Node{}
	err = cur.All(ctx, &results)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return results, nil
}
