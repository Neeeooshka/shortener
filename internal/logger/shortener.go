package logger

import (
	"net/http"
	"time"
)

type (
	Logger interface {
		Log(RequestData, ResponseData)
	}

	RequestData struct {
		URI      string
		Method   string
		Duration time.Duration
	}

	ResponseData struct {
		Status int
		Size   int
	}

	logging struct {
		http.ResponseWriter
		responseData *ResponseData
	}
)

func (r *logging) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size
	return size, err
}

func (r *logging) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode
}

func IncludeLogger(h http.HandlerFunc, l Logger) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &ResponseData{
			Status: 0,
			Size:   0,
		}

		lw := logging{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		requestData := &RequestData{
			URI:      r.RequestURI,
			Method:   r.Method,
			Duration: time.Since(start),
		}
		l.Log(*requestData, *responseData)
	}
	return logFn
}
