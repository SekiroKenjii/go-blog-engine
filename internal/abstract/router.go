package abstract

import "github.com/gin-gonic/gin"

type IRouter interface {
	Engine() *gin.Engine
	SetupRoutes()
}
