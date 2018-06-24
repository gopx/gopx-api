package handler

import "net/http"

// PackagesGET returns the list of all public packages.
// Request: GET /packages?page=1&limit=10&sort=downloads&order=asc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func PackagesGET(w http.ResponseWriter, r *http.Request) {

}

// SinglePackageGET returns the info about a single package.
// Request: GET /packages/:packageName
func SinglePackageGET(w http.ResponseWriter, r *http.Request) {

}

// SinglePackageDELETE deletes a whole package or a specific version.
// It requires user authentication.
// Request: DELETE /packages/:packageName
// Specific version: DELETE /packages/:packageName?v=1.0.1
func SinglePackageDELETE(w http.ResponseWriter, r *http.Request) {

}

// SearchGET performs a query search among all public packages.
// Reference: https://developer.github.com/v3/search/#search-repositories
// Request: GET /search?q=websocket+created:>2017-01-01&page=1&limit=10&sort=downloads&order=desc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func SearchGET(w http.ResponseWriter, r *http.Request) {

}

// DownloadsGET returns the list of download counts of all public packages.
// Request: GET /downloads?page=1&limit=10&sort=downloads&order=asc
// Only total downloads through registry: GET /downloads?totals=1
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func DownloadsGET(w http.ResponseWriter, r *http.Request) {

}

// SinglePackageDownloadsGET returns the download counts of a single package.
// Request: GET /downloads/:packageName
func SinglePackageDownloadsGET(w http.ResponseWriter, r *http.Request) {

}

// VersionsGET returns the list of version histories of all public packages.
// Request: GET /versions?page=1&limit=10&sort=downloads&order=asc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
func VersionsGET(w http.ResponseWriter, r *http.Request) {

}

// SinglePackageVersionsGET returns the version histories of a single package.
// Request: GET /versions/:packageName
func SinglePackageVersionsGET(w http.ResponseWriter, r *http.Request) {

}
