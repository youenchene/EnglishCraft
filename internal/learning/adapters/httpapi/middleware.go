package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

type principalContextKey struct{}

func PrincipalFromContext(ctx context.Context) (identity.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(identity.Principal)
	return principal, ok
}

func Authenticate(authenticate *identity.Authenticate, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		principal, err := authenticate.Execute(request.Context(), bearerToken(request.Header.Get("Authorization")))
		if err != nil {
			response.Header().Set("WWW-Authenticate", "Bearer")
			writeError(response, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
			return
		}
		ctx := context.WithValue(request.Context(), principalContextKey{}, principal)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		response.Header().Set("Referrer-Policy", "no-referrer")
		response.Header().Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(self)")
		response.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; media-src 'self' blob:; connect-src 'self' https://api.openai.com https://api.elevenlabs.io https://*.elevenlabs.io wss://*.elevenlabs.io https://*.livekit.cloud wss://*.livekit.cloud; worker-src 'self'; manifest-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(response, request)
	})
}

func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("recovered HTTP panic", "error", recovered, "stack", string(debug.Stack()))
				writeError(response, http.StatusInternalServerError, "internal_error", "Something went wrong. Please try again.")
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(response, request)
		logger.Info("HTTP request", "method", request.Method, "path", request.URL.Path, "duration_ms", time.Since(startedAt).Milliseconds())
	})
}
