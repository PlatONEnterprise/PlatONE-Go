package engine

import "github.com/gin-gonic/gin"

type Engine struct {
	*gin.Engine
}

var Default *Engine

func init() {
	Default = New()
}

func New() *Engine {
	return &Engine{Engine: gin.New()}
}

// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (this *Engine) Run(addr string) {
	this.Engine.Run(addr)

	panic("failed to web engine start.")
}

