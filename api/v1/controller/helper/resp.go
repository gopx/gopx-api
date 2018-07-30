package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-common/log"
)

func setBasicHeaders(headers http.Header) {
	headers.Set("Server", "GoPx.io")
	headers.Set("Access-Control-Expose-Headers", "Content-Length, Server, Date, Status")
	headers.Set("Access-Control-Allow-Origin", "*")
}

// WriteResponse writes the JSON response to the client.
func WriteResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Error("Error: %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	headers := w.Header()
	setBasicHeaders(headers)

	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(bytes)))
	headers.Set("Status", fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)))

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
