package v1

import (
	"github.com/gorilla/mux"
	"gopx.io/gopx-api/api/v1/handler"
)

// RegisterRoutes registers the routes for API version v1.
func RegisterRoutes(r *mux.Router) {
	registerUserRoutes(r)
	registerPackageRoutes(r)
}

func registerUserRoutes(r *mux.Router) {
	r.Path("/users").
		Methods("GET").
		HandlerFunc(handler.UsersGET)

	r.Path("/users/{username}").
		Methods("GET").
		HandlerFunc(handler.SingleUserGET)

	r.Path("/user").
		Methods("GET").
		HandlerFunc(handler.CurrentUserGET)

	r.Path("/user").
		Methods("PATCH").
		HandlerFunc(handler.CurrentUserPATCH)

	r.Path("/users/{username}/packages").
		Methods("GET").
		HandlerFunc(handler.SingleUserPackagesGET)

	r.Path("/user/packages").
		Methods("GET").
		HandlerFunc(handler.CurrentUserPackagesGET)

	r.Path("/user/packages").
		Methods("POST").
		HandlerFunc(handler.CurrentUserPackagesPOST)
}

func registerPackageRoutes(r *mux.Router) {
	r.Path("/packages").
		Methods("GET").
		HandlerFunc(handler.PackagesGET)

	r.Path("/packages/{packageName}").
		Methods("GET").
		HandlerFunc(handler.SinglePackageGET)

	r.Path("/packages/{packageName}").
		Methods("DELETE").
		HandlerFunc(handler.SinglePackageDELETE)

	r.Path("/packages/{packageName}/files").
		Methods("GET").
		HandlerFunc(handler.SinglePackageFilesGET)

	r.Path("/search").
		Methods("GET").
		HandlerFunc(handler.SearchGET)

	r.Path("/downloads").
		Methods("GET").
		HandlerFunc(handler.DownloadsGET)

	r.Path("/downloads/{packageName}").
		Methods("GET").
		HandlerFunc(handler.SinglePackageDownloadsGET)

	r.Path("/versions").
		Methods("GET").
		HandlerFunc(handler.VersionsGET)

	r.Path("/versions/{packageName}").
		Methods("GET").
		HandlerFunc(handler.SinglePackageVersionsGET)
}
