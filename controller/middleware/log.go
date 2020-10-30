package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type statusCodeLogger struct {
	http.ResponseWriter
	statusCode int
}

func newStatusCodeLogger(rw http.ResponseWriter) *statusCodeLogger {
	return &statusCodeLogger{ResponseWriter: rw, statusCode: http.StatusOK}
}

func (sl *statusCodeLogger) WriteHeader(code int) {
	sl.statusCode = code
	sl.ResponseWriter.WriteHeader(code)
}

// RequestTimeLoggingHandler logs the time a request needs to be processed.
// This handler should be add at the beginning of a handler chain.
// Every request will be logged with the Trace logging level.
func RequestTimeLoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sl := newStatusCodeLogger(rw)
		requestBeginn := time.Now()
		next.ServeHTTP(sl, r)
		requestDuration := time.Since(requestBeginn)
		logPackage.WithFields(logrus.Fields{
			"RequestTime":   requestDuration.Milliseconds(),
			"RequestPath":   r.URL.String(),
			"RequestMethod": r.Method,
			"StatusCode":    sl.statusCode,
		}).Tracef("[%s] %q %v\n", r.Method, r.URL.String(), requestDuration)
	})
}
