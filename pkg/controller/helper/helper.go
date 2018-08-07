package helper

import (
	"net/http"

	"gopx.io/gopx-common/log"
)

// WriteRespData writes the input data to the http request.
func WriteRespData(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		log.Error("Error %s", err)
	}
}
