package error

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Error handles request which causes any error.
func Error(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	writeResponse(w, statusCode, message)
}

// Error401 handles unauthorized request.
func Error401(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnauthorized, "Requires authentication")
}

// Error403 handles forbidden request.
func Error403(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusForbidden, "Requires permissions")
}

// Error404 handles request for non-existing resources.
func Error404(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusNotFound, "Not Found")
}

// Error405 handles request with not allowed http method.
func Error405(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusMethodNotAllowed, "Method Not Allowed")
}

// Error500 handles request with internal server error.
func Error500(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusInternalServerError, "Internal Server Error")
}

// DecodeErrorMessage extracts the message part of error encoded response data.
func DecodeErrorMessage(errData []byte) (message string, err error) {
	r := bytes.NewReader(errData)

	var oErrResp errorResponse
	err = json.NewDecoder(r).Decode(&oErrResp)
	if err != nil {
		return
	}

	message = oErrResp.Message

	return
}
