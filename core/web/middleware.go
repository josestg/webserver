package web

// Middleware is a function designed to run some code before and/or after another handler.
// It is designed to remove boilerplate or other concerns not direct to any given handler.
type Middleware func(handler Handler) Handler

// wrapMiddleware creates a new handler by wrapping middleware around a final handler.
// The middleware Handlers will be executed by request in the order they are provided.
func wrapMiddleware(middlewares []Middleware, handler Handler) Handler {

	// Loop backward through the middleware invoking each one.
	// Replace the handler with the new wrapped handler.
	// Looping backward ensures that the first middleware of the slice
	// is the first to be executed by request.
	for i := len(middlewares) - 1; i >= 0; i-- {
		h := middlewares[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
