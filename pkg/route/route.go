package route

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gopx.io/gopx-api/api/v1"
	"gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-api/pkg/log"
)

// Router registers the API specific routes.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	r.Use(loggingMiddleware)

	s1 := r.PathPrefix("/v1").Subrouter()
	v1.RegisterRoutes(s1)

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	error.Error404(w, r)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	error.Error405(w, r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("%s %s", strings.ToUpper(r.Method), r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
