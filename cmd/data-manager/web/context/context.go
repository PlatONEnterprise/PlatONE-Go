package context

import (
	dbCtx "data-manager/db/context"
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
	DBCtx *dbCtx.Context
}

func New() *Context {
	return &Context{}
}

func (this *Context) SetContext(ctx *gin.Context) {
	this.Context = ctx
}

func (this *Context) SetDBContext(ctx *dbCtx.Context) {
	this.DBCtx = ctx
}
