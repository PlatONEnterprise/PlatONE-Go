package context

import (
	"data-manager/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type Context struct {
	db *db.DB
}

func New(db *db.DB) *Context {
	return &Context{db: db}
}

func (this *Context) Collection(name string) *mongo.Collection {
	return this.db.Collection(name)
}
