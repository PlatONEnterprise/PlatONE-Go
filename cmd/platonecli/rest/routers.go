package rest

import (
	"github.com/gin-gonic/gin"
)

func StartServer(endPoint string) {
	r := genRestRouters()
	_ = r.Run(endPoint)
}

func genRestRouters() *gin.Engine {
	router := gin.New()

	// todo: custom middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	registerRouters(router)

	return router
}

func registerRouters(r *gin.Engine) {
	registerAccountRouters(r)
	registerCnsRouters(r)
	registerContractRouters(r)
	registerFwRouters(r)
	registerNodeRouters(r)
	registerRoleRouters(r)
	registerSysConfigRouters(r)

	/// registerRpcAPIs(r)
}
