package middleware

import "net/http"

// CORS adds cross-origin headers. Allow origins are derived from the
// CORS_ORIGIN environment variable (comma-separated). When CORS_ORIGIN is
// empty or set to "*", all origins are permitted (development convenience).
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := splitOrigins(allowedOrigins)
	allowAll := len(origins) == 0 || (len(origins) == 1 && origins[0] == "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && originMatches(origin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func splitOrigins(raw string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == ',' {
			s := raw[start:i]
			if s != "" {
				out = append(out, s)
			}
			start = i + 1
		}
	}
	return out
}

func originMatches(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}
