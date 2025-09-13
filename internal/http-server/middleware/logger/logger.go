package logger

import (
	"log/slog"
	"net/http"
	"time"
)

type responseData struct {
	status       int
	bytesWritten int
}

type responseWriterWithLogging struct {
	http.ResponseWriter
	responseData *responseData
}

func (r responseWriterWithLogging) Write(b []byte) (int, error) {
	bytesWritten, err := r.ResponseWriter.Write(b)
	r.responseData.bytesWritten += bytesWritten
	return bytesWritten, err
}
func (r responseWriterWithLogging) WriteHeader(status int) {
	r.responseData.status = status
	r.ResponseWriter.WriteHeader(status)
}

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("Logger middleware is enabled")
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			resData := &responseData{}
			rwl := responseWriterWithLogging{
				ResponseWriter: w,
				responseData:   resData,
			}

			start := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", resData.status),
					slog.Int("bytes", resData.bytesWritten),
					slog.Duration("time", time.Since(start)),
				)
			}()

			next.ServeHTTP(rwl, r)
		}

		return http.HandlerFunc(fn)
	}
}
