package util

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(statusCode int, obj interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(obj)
}
