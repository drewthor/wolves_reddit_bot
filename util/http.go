package util

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(statusCode int, obj interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, err := json.Marshal(obj)
	if err != nil {
		return
	}

	w.Write(b)
}
