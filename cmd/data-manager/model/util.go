package model

import "go.mongodb.org/mongo-driver/mongo/options"

func buildOptionsByQuery(pageIndex, pageSize int64) *options.FindOptions {
	findOps := options.Find()
	findOps.SetSkip((pageIndex - 1) * pageSize)
	findOps.SetLimit(pageSize)

	return findOps
}
