package service

import "net/http"

type Middleware struct {
	auth bool
}

// NewMiddleware is a struct that has a ServeHTTP method
func NewMiddleware() *Middleware {
	return &Middleware{true}
}

// The middleware handler
func (l *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	key := req.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")
	if key == "" {
		http.Error(w, "Failed to find token", http.StatusInternalServerError)
		return
	}

	_, err := REDIS.Get(key).Result()
	if err != nil {
		http.Error(w, "Not a valid token", http.StatusInternalServerError)
		return
	}
	next(w, req)
}
