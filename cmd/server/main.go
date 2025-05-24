package main

import (
	"github.com/SekiroKenjii/go-blog-engine/core"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"

	_ "github.com/SekiroKenjii/go-blog-engine/docs"
)

// @title			Blog Engine API
// @version			1.0
// @description		A fully-featured blogging platform using Golang to explore scalable back-end design and concurrency patterns.
// @termsOfService	https://thuongvo.dev/terms
// @contact.name	Thuong Vo
// @contact.url		https://thuongvo.dev/support
// @contact.email	thuongvo.dev.99@gmail.com
// @license.name	Apache 2.0
// @license.url		http://www.apache.org/licenses/LICENSE-2.0.html
// @host			thuongvo.dev:8080
// @BasePath		/api/v1
// @schemas			http https
func main() {
	app := core.Bootstrap()
	httpSrv := app.BuildHttpServer()

	done := make(chan bool, 1)

	go app.Shutdown(httpSrv, done)

	app.Run(httpSrv)

	// Wait for the shutdown process to complete
	<-done

	logger.Info("Application shutdown complete.")
}
