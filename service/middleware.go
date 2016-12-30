package service

import "net/http"

// The middleware handler
func AuthMiddleware(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
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
