package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/terjelafton/platform/libs/logger"
)

// CorrelationID generates a UUID v7 per request and stores it in the context.
// Must run before other middleware so all downstream logs include the ID.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.Must(uuid.NewV7()).String()
		ctx := logger.WithCorrelationID(r.Context(), id)
		w.Header().Set("X-Correlation-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
