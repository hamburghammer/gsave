package middleware

import "net/http"

// AuthMiddleware is a struct to hold a array of valid tokens.
type AuthMiddleware struct {
	tokens []string
}

// NewAuthMiddleware is a constructor for the AuthMiddleware struct.
func NewAuthMiddleware(tokens []string) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

// AuthHandler implements the handling of a request and checks if it is authorized.
// It checks if the 'Token' header is set and if the token is valid.
// If the header is missing it will return a http.StatusBadRequest and if the token isn't
// valid it will return a http.StatusUnauthorized status code.
func (am *AuthMiddleware) AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		tokens := r.Header["Token"]
		if len(tokens) < 1 {
			http.Error(rw, "Missing 'Token' header", http.StatusBadRequest)
			return
		}
		token := tokens[0]
		if !am.isValid(token) {
			http.Error(rw, "The token is not valid", http.StatusUnauthorized)
			logPackage.Warnf("Login attempt with wrong token: '%s' from ip: '%s'\n", token, r.RemoteAddr)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func (am *AuthMiddleware) isValid(token string) bool {
	for _, authToken := range am.tokens {
		if token == authToken {
			return true
		}
	}
	return false
}
