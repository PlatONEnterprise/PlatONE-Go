package db

import (
	"context"
	"data-manager/config"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

var DefaultDB *DB = newDB()

func init() {
	logrus.Infof("db uri:%s", config.Config.DBConf.Uri())
	optionUri := options.Client().ApplyURI(config.Config.DBConf.Uri())
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cli, err := mongo.Connect(ctx, optionUri)
	if nil != err {
		panic(err)
	}

	DefaultDB.client = cli
	DefaultDB.db = cli.Database(config.Config.DBConf.DBName)

	if err = DefaultDB.Ping(); nil != err {
		panic(err)
	}

	logrus.Info("db successfully connected and pinged.")
}

type DB struct {
	client *mongo.Client
	db     *mongo.Database
}

func newDB() *DB {
	return &DB{}
}

func (this *DB) Collection(name string) *mongo.Collection {
	return this.db.Collection(name)
}

func (this *DB) Client() *mongo.Client {
	return this.client
}

//func (this *DB) Database() *mongo.Database {
//	return this.db
//}

func (this *DB) Ping() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err := this.client.Ping(ctx, readpref.Primary()); nil != err {
		return err
	}

	return nil
}
