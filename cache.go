package main

import "net/http"

func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
		next.ServeHTTP(w, r)
	})
}
