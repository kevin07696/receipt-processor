package receipt

import "net/http"

func HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This just checks that the application is still running
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}