package error

import (
	"io"
	"net/http"

	"gopx.io/gopx-api/pkg/utils"
)

// Error403 handles forbidden request.
func Error403(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	_, err := io.WriteString(w, "403 Forbidden")
	utils.LogWarn(err)
}

// Error404 handles HTTP request on non-existing routes.
func Error404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, err := io.WriteString(w, "404 Not Found")
	utils.LogWarn(err)
}

// Error405 handles not allowed http method.
func Error405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err := io.WriteString(w, "405 Method Not Allowed")
	utils.LogWarn(err)
}

// Error500 handles internal server error.
func Error500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := io.WriteString(w, "500 Internal Server Error")
	utils.LogWarn(err)
}
