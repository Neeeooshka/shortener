package logger

import (
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
	"net/http"
)

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.Log.Debug("got incoming HTTP request",
			zap.Log.String("method", r.Method),
			zap.Log.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
