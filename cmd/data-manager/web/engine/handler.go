package engine

import (
	"data-manager/db"
	dbContext "data-manager/db/context"
	webContext "data-manager/web/context"
	"github.com/gin-gonic/gin"
)

type HandlerFunc func(ctx *webContext.Context)

func NewHandler(handlers ...HandlerFunc) func(*gin.Context) {
	return func(ctx *gin.Context) {
		webCtx := webContext.New()

		webCtx.SetContext(ctx)

		dbCtx := dbContext.New(db.DefaultDB)
		webCtx.SetDBContext(dbCtx)

		for _, h := range handlers {
			h(webCtx)
		}
	}
}
