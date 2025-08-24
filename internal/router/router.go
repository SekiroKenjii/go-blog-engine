package router

import (
	"fmt"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/SekiroKenjii/go-blog-engine/internal/middlewares"
	"github.com/SekiroKenjii/go-blog-engine/pkg/scalar"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter() abstract.IRouter {
	env := config.Instance().Server.Env
	if env == "develop" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()

		return &Router{
			engine: gin.Default(),
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	return &Router{
		engine: r,
	}
}

// Configure implements IRouter.
func (r *Router) Configure() {
	r.addMiddlewares()
	r.addOpenAPI()
	r.addAPIRoutes()
}

// addMiddlewares adds the necessary middlewares to the gin engine.
func (r *Router) addMiddlewares() {
	r.engine.Use(middlewares.ErrorHandler())
	r.engine.Use(middlewares.Cors())
	r.engine.Use(middlewares.RateLimitExcludingPaths("/api/v1/auth"))
}

// addOpenAPI adds the OpenAPI documentation routes to the gin engine.
func (r *Router) addOpenAPI() {
	r.engine.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.engine.GET("/docs/scalar", func(c *gin.Context) {
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

// Engine implements IRouter.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
