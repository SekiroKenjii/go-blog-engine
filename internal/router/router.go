package router

import (
	"fmt"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/middlewares"
	"github.com/SekiroKenjii/go-blog-engine/pkg/scalar"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	instance *Router
	once     sync.Once
)

type Router struct {
	Engine *gin.Engine
}

func Instance() abstract.IRouter {
	once.Do(func() {
		instance = newRouter()
	})

	return instance
}

func newRouter() *Router {
	env := config.Instance().Server.Env
	if env == "develop" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()

		return &Router{
			Engine: gin.Default(),
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	return &Router{
		Engine: r,
	}
}

func (r *Router) SetupRoutes() {
	r.addMiddlewares()
	r.addOpenAPI()
	r.addAPIRoutes()
}

func (r *Router) GetEngine() *gin.Engine {
	return r.Engine
}

func (r *Router) addMiddlewares() {
	r.Engine.Use(middlewares.ErrorHandler())
	r.Engine.Use(middlewares.Cors())
	r.Engine.Use(middlewares.RateLimit())
	r.Engine.Use(middlewares.Auth())
}

func (r *Router) addOpenAPI() {
	r.Engine.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Engine.GET("/docs/scalar", func(c *gin.Context) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "./docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Blog Engine API",
			},
			DarkMode: true,
		})
		if err != nil {
			fmt.Printf("%v", err)
		}

		fmt.Fprintln(c.Writer, htmlContent)
	})
}
