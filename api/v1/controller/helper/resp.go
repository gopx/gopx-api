package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-api/pkg/controller/helper"
	"gopx.io/gopx-common/log"
)

func setBasicHeaders(headers http.Header) {
	headers.Set("Server", "GoPx.io")
	headers.Set("Access-Control-Expose-Headers", "Content-Length, Server, Date, Status")
	headers.Set("Access-Control-Allow-Origin", "*")
}

// WriteResponse writes JSON data to the client with the specified status code.
func WriteResponse(w http.ResponseWriter, r *http.Request, data []byte, statusCode int) {
	headers := w.Header()
	setBasicHeaders(headers)

	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(data)))
	headers.Set("Status", fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)))

	w.WriteHeader(statusCode)
	if data != nil {
		helper.WriteRespData(w, data)
	}
}

// WriteResponseValue writes the input golang value in the form of JSON encoding
// to the client with the specified status code.
func WriteResponseValue(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	buff := bytes.Buffer{}
	enc := json.NewEncoder(&buff)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	err := enc.Encode(data)
	if err != nil {
		log.Error("Error: %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	WriteResponse(w, r, buff.Bytes(), statusCode)
}

// WriteResponseValueOK writes the input golang value in the form of JSON encoding
// to the client with "200 OK" status.
func WriteResponseValueOK(w http.ResponseWriter, r *http.Request, data interface{}) {
	WriteResponseValue(w, r, data, http.StatusOK)
}
