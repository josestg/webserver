package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	// If the context is missing this value, this is a serious problem, because App Handle is never executed.
	// Therefore, a request must be made to the service to be shutdown gracefully .
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("web value missing from context.")
	}

	// Set the status code for the request logger middleware.
	v.StatusCode = statusCode

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

// RespondError sends an error response back to the client.
func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	// If the error was of the type *Error (or trusted error),
	// the handler has the specific status code and error to return.
	if webErr, ok := errors.Cause(err).(*Error); ok {
		response := ErrorResponse{
			Error:  webErr.Error(),
			Fields: webErr.Fields,
		}

		if err := Respond(ctx, w, response, webErr.Status); err != nil {
			return err
		}
	}

	// If not a trusted error, the handler sent any arbitrary error value so use 500.
	response := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}

	if err := Respond(ctx, w, response, http.StatusInternalServerError); err != nil {
		return err
	}
	return nil
}
