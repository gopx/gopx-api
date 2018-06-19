package helper

import (
	"encoding/json"
	"net/http"

	"gopx.io/gopx-api/pkg/log"
)

// APIResponse represents response for Registry API.
type APIResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func writeAPIResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	apiRes := APIResponse{
		Status:  status,
		Message: message,
	}

	bytes, err := json.Marshal(apiRes)
	if err != nil {
		Error500(w, r)
		log.Error("Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(bytes)
	if err != nil {
		log.Error("Error: %s", err)
	}
}
