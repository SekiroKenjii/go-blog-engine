package abstract

import "github.com/gin-gonic/gin"

type IRouter interface {
	SetupRoutes()
	GetEngine() *gin.Engine
}
