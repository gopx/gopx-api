package error

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gopx.io/gopx-common/log"
)

type errorResponse struct {
	Message string `json:"message"`
}

func setBasicHeaders(headers http.Header) {
	headers.Set("Server", "GoPx.io")
	headers.Set("Access-Control-Expose-Headers", "Content-Length, Server, Date, Status")
	headers.Set("Access-Control-Allow-Origin", "*")
}

func responseJSON(message string) ([]byte, error) {
	resp := errorResponse{Message: message}
	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	bytes, err := responseJSON(message)
	if err != nil {
		log.Error("Error: %s", err)
		bytes = []byte(fmt.Sprintf("\"%s\"", message))
	}

	headers := w.Header()
	setBasicHeaders(headers)

	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(bytes)))
	headers.Set("Status", fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)))

	w.WriteHeader(statusCode)
	w.Write(bytes)
}
