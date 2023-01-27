package web

import (
	"net/http"

	fbauth "firebase.google.com/go/v4/auth"
	"golang.org/x/exp/slog"
)

func AuthenticateMiddleware(log *slog.Logger, authClient *fbauth.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("__session")
			if err != nil {
				InvalidAuthenticationResponse(log, w, r)
				return
			}

			_, err = authClient.VerifySessionCookieAndCheckRevoked(r.Context(), cookie.Value)
			if err != nil {
				InvalidAuthenticationResponse(log, w, r)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
