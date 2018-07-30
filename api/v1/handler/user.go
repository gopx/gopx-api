package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/helper"
	"gopx.io/gopx-api/api/v1/controller/user"
	"gopx.io/gopx-api/api/v1/types"
	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-common/log"
	"gopx.io/gopx-common/misc"
)

// UsersGET returns the list of all users.
// Request: GET /users?page=1&per_page=100&sort=packages&order=asc
// Sorting can be performed on:
// 1. packages
// 2. joined
// 4. username
// 5. id
func UsersGET(w http.ResponseWriter, r *http.Request) {
	qParams := r.URL.Query()

	var (
		pageStr    = qParams.Get("page")
		perPageStr = qParams.Get("per_page")
		sort       = qParams.Get("sort")
		order      = qParams.Get("order")
	)

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		page = 1
	}

	perPage, err := strconv.ParseUint(perPageStr, 10, 64)
	if err != nil {
		perPage = uint64(constants.UsersQueryPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	userRows, err := user.SearchUser(nil, &pc, &sc)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	users := make([]*types.User, len(userRows))

	for i, ur := range userRows {
		users[i] = &types.User{
			Username:      ur.Username,
			ID:            ur.ID,
			Name:          ur.Name,
			Email:         misc.TerOpt(ur.IsPublicEmail, ur.Email, "").(string),
			JoinedAt:      ur.JoinedAt,
			AvatarURL:     ur.Avatar,
			Blog:          ur.URL,
			Organization:  ur.Organization,
			Location:      ur.Location,
			PackagesCount: ur.PackagesCount,
			Social: types.UserSocialAccounts{
				Github:        ur.Github,
				Twitter:       ur.Twitter,
				StackOverflow: ur.StackOverflow,
				LinkedIn:      ur.LinkedIn,
			},
		}
	}

	helper.WriteResponse(w, r, users)
}

// SingleUserGET returns the public information about a single user.
// Request: GET /users/:username
func SingleUserGET(w http.ResponseWriter, r *http.Request) {
	inputUsername := mux.Vars(r)["username"]
	whereClause := "username = ?"
	sortBy := "id ASC"
	limit := "1"

	userRows, err := user.QueryUser(whereClause, sortBy, limit, "", inputUsername)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if len(userRows) == 0 {
		errorCtrl.Error404(w, r)
		return
	}

	ur := userRows[0]
	user := &types.User{
		Username:      ur.Username,
		ID:            ur.ID,
		Name:          ur.Name,
		Email:         misc.TerOpt(ur.IsPublicEmail, ur.Email, "").(string),
		JoinedAt:      ur.JoinedAt,
		AvatarURL:     ur.Avatar,
		Blog:          ur.URL,
		Organization:  ur.Organization,
		Location:      ur.Location,
		PackagesCount: ur.PackagesCount,
		Social: types.UserSocialAccounts{
			Github:        ur.Github,
			Twitter:       ur.Twitter,
			StackOverflow: ur.StackOverflow,
			LinkedIn:      ur.LinkedIn,
		},
	}

	helper.WriteResponse(w, r, user)
}

// CurrentUserGET returns the public and private information of the authenticated user.
// Request: GET /user
func CurrentUserGET(w http.ResponseWriter, r *http.Request) {

}

// CurrentUserPATCH updates the authenticated user.
// Request: PATCH /user
func CurrentUserPATCH(w http.ResponseWriter, r *http.Request) {

}

// SingleUserPackagesGET returns the list of public packages of a single user.
// Request: GET /users/:username/packages?page=1&limit=10&sort=downloads&order=desc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func SingleUserPackagesGET(w http.ResponseWriter, r *http.Request) {

}

// CurrentUserPackagesGET returns the list of all packages of the authenticated user.
// Request: GET /user/packages?page=1&limit=10&sort=downloads&order=desc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func CurrentUserPackagesGET(w http.ResponseWriter, r *http.Request) {

}

// CurrentUserPackagesPOST register a new package to the authenticated user.
// Request: POST /user/packages
func CurrentUserPackagesPOST(w http.ResponseWriter, r *http.Request) {

}
