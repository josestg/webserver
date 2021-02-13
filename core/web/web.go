package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// Handler is a type that handles an http request within our own little mini web framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what our configures our context object
// for each of http handlers. Feel free to add any configuration data/logic on this app struct.
type App struct {
	*httptreemux.ContextMux
	shutdown    chan os.Signal
	middlewares []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	app := App{
		ContextMux:  httptreemux.NewContextMux(),
		shutdown:    shutdown,
		middlewares: mw,
	}

	return &app
}

// SignalShutdown is used to gracefully shutdown the app when an integrity issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle ...
func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {

	// First wrap the handler specific (local) middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.middlewares, handler)

	handlerAdapter := func(w http.ResponseWriter, r *http.Request) {

		// Set the context with the required values to process the request.
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)

		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}
	a.ContextMux.Handle(method, path, handlerAdapter)
}
