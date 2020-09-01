package engine

import (
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/db"
	dbContext "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/db/context"
	webContext "github.com/PlatONEnetwork/PlatONE-Go/cmd/data-manager/web/context"
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
