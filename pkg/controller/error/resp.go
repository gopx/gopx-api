package error

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gopx.io/gopx-api/pkg/controller/helper"
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
	buff := bytes.Buffer{}
	enc := json.NewEncoder(&buff)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	resp := errorResponse{Message: message}
	err := enc.Encode(resp)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	bytes, err := responseJSON(message)
	if err != nil {
		log.Error("Error: %s", err)
		bytes = []byte(fmt.Sprintf("\"%s\"\n", message))
	}

	headers := w.Header()
	setBasicHeaders(headers)

	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(bytes)))
	headers.Set("Status", fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)))

	w.WriteHeader(statusCode)
	helper.WriteRespData(w, bytes)
}
