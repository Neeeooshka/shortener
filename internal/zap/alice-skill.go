package zap

import (
	"net/http"
)

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Debug("got incoming HTTP request",
			Log.String("method", r.Method),
			Log.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
