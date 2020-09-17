package rest

import "github.com/gin-gonic/gin"

func registerRpcAPIs(r *gin.Engine) {
	r.GET("/blockNum", getBlockNumHandler)
}

// ===================== RPC =========================
func getBlockNumHandler(ctx *gin.Context) {}
