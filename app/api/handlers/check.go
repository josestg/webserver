package handlers

import (
	"context"
	"log"
	"net/http"
)

// Check checks handler
type Check struct {
	logger *log.Logger
}

func (c Check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}
