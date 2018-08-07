package v1

import (
	"github.com/gorilla/mux"
	"gopx.io/gopx-api/api/v1/handler"
)

// RegisterRoutes registers the routes for API version v1.
func RegisterRoutes(r *mux.Router) {
	r.Path("/users").
		Methods("GET").
		HandlerFunc(handler.UsersGET)

	r.Path("/users/{username}").
		Methods("GET").
		HandlerFunc(handler.SingleUserGET)

	r.Path("/users/{username}/packages").
		Methods("GET").
		HandlerFunc(handler.SingleUserPackagesGET)

	r.Path("/user").
		Methods("GET").
		HandlerFunc(handler.CurrentUserGET)

	r.Path("/user").
		Methods("PATCH").
		HandlerFunc(handler.CurrentUserPATCH)

	r.Path("/user/packages").
		Methods("GET").
		HandlerFunc(handler.CurrentUserPackagesGET)

	r.Path("/user/packages").
		Methods("POST").
		HandlerFunc(handler.CurrentUserPackagesPOST)

	r.Path("/user/packages/{packageName}").
		Methods("DELETE").
		HandlerFunc(handler.CurrentUserPackagesDELETE)

	r.Path("/packages").
		Methods("GET").
		HandlerFunc(handler.PackagesGET)

	r.Path("/packages/{packageName}").
		Methods("GET").
		HandlerFunc(handler.SinglePackageGET)

	r.Path("/packages/{packageName}/readme").
		Methods("GET").
		HandlerFunc(handler.SinglePackageReadmeGET)

	r.Path("/downloads").
		Methods("GET").
		HandlerFunc(handler.DownloadsGET)

	r.Path("/versions/{packageName}").
		Methods("GET").
		HandlerFunc(handler.SinglePackageVersionsGET)

	r.Path("/search/users").
		Methods("GET").
		HandlerFunc(handler.SearchUsersGET)

	r.Path("/search/packages").
		Methods("GET").
		HandlerFunc(handler.SearchPackagesGET)
}
