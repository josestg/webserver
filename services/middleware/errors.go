package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/josestg/webserver/core/web"
)

// Errors handles error coming out of the call chain.
// It detects normal application errors with are used to respond to the client in a uniform way.
// Unexpected errors or untrusted errors (status >= 500) are logged.
func Errors(logger *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	errorsMiddleware := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context.")
			}

			// Run the handler chain and catch anya propagated error.
			if err := handler(ctx, w, r); err != nil {
				logger.Printf("%s: ERROR: %v", v.TraceID, err)

				// Respond to the error.
				if err := web.RespondError(ctx, w, err); err != nil {
					return err
				}

				// If we receive the shutdown error we need to return it back
				// to the base handler to shutdown the service.
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}

		return h

	}
	return errorsMiddleware
}
