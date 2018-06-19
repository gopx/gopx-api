package route

import (
	"github.com/gorilla/mux"
)

func GoPXAPIRouter() *mux.Router {
	return mux.NewRouter()
}
