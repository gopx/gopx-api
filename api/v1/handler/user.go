package handler

import "net/http"

// UsersGET returns the list of all users.
// Request: GET /users?page=1&limit=10&sort=packages&order=asc
// Sorting can be performed on:
// 1. packages
// 2. joined
// 3. updated
// 4. username
func UsersGET(w http.ResponseWriter, r *http.Request) {

}

// SingleUserGET returns the public information about a single user.
// Request: GET /users/:username
func SingleUserGET(w http.ResponseWriter, r *http.Request) {

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
