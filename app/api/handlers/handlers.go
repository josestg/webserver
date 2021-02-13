package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/josestg/webserver/core/web"
	"github.com/josestg/webserver/services/middleware"
)

// API constructs an http.Handler with all application routes defined.
func API(logger *log.Logger, shutdown chan os.Signal) *web.App {
	app := web.NewApp(
		shutdown,
		middleware.Logger(logger),
		middleware.Errors(logger),
		middleware.Panics(logger),
	)

	check := Check{logger: logger}
	app.Handle(http.MethodGet, "/readiness", check.readiness)

	return app
}
