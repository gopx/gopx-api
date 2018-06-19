package controller

import (
	"io"
	"net/http"

	"gopx.io/gopx-api/pkg/utils"
)

// Error404 handles HTTP request on non-existing routes.
func Error404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, err := io.WriteString(w, "404 Not Found")
	utils.LogWarn(err)
}
