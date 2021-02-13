package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/pkg/errors"

	"github.com/josestg/webserver/core/web"
)

// Panics recovers from panics and converts the panic to an error
// so it is reported in Metrics and handled in Errors.
func Panics(logger *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	panicsMiddleware := func(handler web.Handler) web.Handler {

		//  Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// If the context is missing this value, request the service to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context.")
			}

			// Defer a function to recover from a panic and set the error return variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					err = errors.Errorf("panic: %v", rec)

					// Log the Go stack trace for this panic's goroutine.
					logger.Printf("%s: PANIC:  %\n%s", v.TraceID, debug.Stack())
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return panicsMiddleware
}
