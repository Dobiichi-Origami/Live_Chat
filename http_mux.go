package liveChat

import (
	"github.com/gin-gonic/gin"
)

var httpServer *gin.Engine

const defaultRouterConfigPath = "./router.json"

var RouterConfigPath = defaultRouterConfigPath

func InitiateHttpServer(configPath string) {
	httpServer = gin.Default()
}
