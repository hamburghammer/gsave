package middleware

import "net/http"

// PanicRecoverHandler is a Handler to recover from any panic that happend down the handler chain.
// If a panic occurs it will generate an error log.
func PanicRecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logPackage.Errorf("Recovered from a panic: %+v\n", err)
				http.Error(w, "Something went wrong.", 500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
