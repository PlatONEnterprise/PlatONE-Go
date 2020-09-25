package model

import (
	"context"
	dbCtx "data-manager/db/context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

const (
	collectionNameBlocks = "blocks"
)

var DefaultBlock = newBlock()

type block struct {
}

func newBlock() *block {
	return new(block)
}

func (this *block) InsertBlock(c *dbCtx.Context, b *Block) error {
	collection := c.Collection(collectionNameBlocks)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	_, err := collection.InsertOne(ctx, b)
	if nil != err {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (this *block) LatestHeight(c *dbCtx.Context) (uint64, error) {
	block, err := this.LatestBlock(c)
	if nil != err {
		logrus.Errorln(err)
		return 0, err
	}

	return block.Height, nil
}

func (this *block) LatestBlock(c *dbCtx.Context) (*Block, error) {
	collection := c.Collection(collectionNameBlocks)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var b Block
	err := collection.FindOne(
		ctx,
		bsonx.Doc{},
		options.FindOne().SetSort(bson.D{{"height", -1}})).Decode(&b)
	if nil != err && mongo.ErrNoDocuments != err {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("FindLatestBlock in db:%+v", b)
	return &b, nil
}

func (this *block) BlockByHash(c *dbCtx.Context, hash string) (*Block, error) {
	filter := bson.M{"hash": hash}
	return this.block(c, filter)
}

func (this *block) BlockByHeight(c *dbCtx.Context, height uint64) (*Block, error) {
	filter := bson.M{"height": height}
	return this.block(c, filter)
}

func (this *block) block(c *dbCtx.Context, filter interface{}) (*Block, error) {
	collection := c.Collection(collectionNameBlocks)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var b Block
	err := collection.FindOne(ctx, filter).Decode(&b)
	if nil != err {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter:", filter, "model block:%+v", b)
	return &b, nil
}

func (this *block) Blocks(c *dbCtx.Context, pageIndex, pageSize int64) ([]*Block, error) {
	collection := c.Collection(collectionNameBlocks)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.D{{}}
	findOps := buildOptionsByQuery(pageIndex, pageSize)
	findOps.SetSort(bson.D{{"height", -1}})

	cur, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*Block, 20)
	err = cur.All(ctx, &results)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return results, nil
}
