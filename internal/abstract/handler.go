package abstract

import "github.com/gin-gonic/gin"

type IHandler interface {
	RegisterRoutes(*gin.RouterGroup)
}
