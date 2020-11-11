package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthHandler(t *testing.T) {
	t.Run("no token provided", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware([]string{"foo"})
		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.Handler(
			authMiddleware.AuthHandler(
				http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					t.Error("Should have bin blocked by the middleware")
				}),
			),
		)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "Missing 'Token' header\n", rr.Body.String())
	})

	t.Run("wrong token provided", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware([]string{"foo"})
		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Token", "bar")

		rr := httptest.NewRecorder()
		handler := http.Handler(
			authMiddleware.AuthHandler(
				http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					t.Error("Should have bin blocked by the middleware")
				}),
			),
		)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Equal(t, "The token is not valid\n", rr.Body.String())
	})

	t.Run("correct token provided", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware([]string{"foo"})
		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Token", "foo")

		rr := httptest.NewRecorder()
		handler := http.Handler(
			authMiddleware.AuthHandler(
				http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					rw.WriteHeader(http.StatusOK)
				}),
			),
		)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

}
