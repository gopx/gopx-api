package error

import (
	"encoding/json"
	"io"
	"net/http"

	"gopx.io/gopx-common/log"
)

type errorResponse struct {
	Message string `json:"message"`
}

func responseJSON(message string) (string, error) {
	resp := errorResponse{Message: message}
	bytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	resp, err := responseJSON(message)
	if err != nil {
		log.Error("Error: %s", err)
		resp = message
	}

	io.WriteString(w, resp)
}
