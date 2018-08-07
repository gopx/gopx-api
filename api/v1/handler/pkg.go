package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gopx.io/gopx-common/str"

	"github.com/gorilla/mux"

	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/helper"
	"gopx.io/gopx-api/api/v1/controller/pkg"
	"gopx.io/gopx-api/api/v1/types"
	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-api/pkg/controller/vcs"
	"gopx.io/gopx-common/log"
)

// PackagesGET returns the list of all public packages.
// Request: GET /packages?page=1&per_page=10&sort=downloads&order=asc
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
// 5. id
func PackagesGET(w http.ResponseWriter, r *http.Request) {
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
		perPage = uint64(constants.PackagesQueryMaxPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	pkgRows, err := pkg.Search(nil, &pc, &sc)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	pkgs := make([]*types.Package, len(pkgRows))

	for i, pr := range pkgRows {
		pkgs[i] = &types.Package{
			Name:             pr.Name,
			ID:               pr.ID,
			Desc:             pr.Description,
			Owner:            pr.OwnerUsername,
			Version:          pr.LatestVersion,
			Downloads:        pr.Downloads,
			PublishedAt:      pr.PublishedAt,
			UpdatedAt:        pr.LastReleasedAt,
			License:          pr.License,
			Homepage:         pr.HomepageURL,
			RepositoryURL:    pr.RepositoryURL,
			DocumentationURL: pr.DocumentationURL,
			BugsURL:          pr.BugsURL,
			Engines: types.Engines{
				Go: pr.EnginesGO,
			},
			Os: listOsNames(pr.OS),
		}
	}

	helper.WriteResponseJSONOk(w, r, pkgs)
}

// SinglePackageGET returns the info about a single package.
// Request: GET /packages/:packageName
func SinglePackageGET(w http.ResponseWriter, r *http.Request) {
	inputPkgName := mux.Vars(r)["packageName"]
	whereClause := "name = ?"
	sortBy := "id ASC"
	limit := "1"

	pkgRows, err := pkg.Query(whereClause, sortBy, limit, "", inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if len(pkgRows) == 0 {
		errorCtrl.Error404(w, r)
		return
	}

	pr := pkgRows[0]
	pkg := &types.Package{
		Name:             pr.Name,
		ID:               pr.ID,
		Desc:             pr.Description,
		Owner:            pr.OwnerUsername,
		Version:          pr.LatestVersion,
		Downloads:        pr.Downloads,
		PublishedAt:      pr.PublishedAt,
		UpdatedAt:        pr.LastReleasedAt,
		License:          pr.License,
		Homepage:         pr.HomepageURL,
		RepositoryURL:    pr.RepositoryURL,
		DocumentationURL: pr.DocumentationURL,
		BugsURL:          pr.BugsURL,
		Engines: types.Engines{
			Go: pr.EnginesGO,
		},
		Os: listOsNames(pr.OS),
	}

	helper.WriteResponseJSONOk(w, r, pkg)
}

// SearchPackagesGET performs a search query among all public packages.
// Request: GET /search/packages?q=websocket+in:name,desc+created:>2017-01-01&sort=downloads,id&order=desc&page=1&per_page=10
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
// 5. id
func SearchPackagesGET(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	var (
		q          = params.Get("q")
		pageStr    = params.Get("page")
		perPageStr = params.Get("per_page")
		sort       = params.Get("sort")
		order      = params.Get("order")
	)

	sq := pkg.SearchQuery{}
	q = strings.TrimSpace(q)
	parts := str.SplitSpace(q)
	re, _ := regexp.Compile("^([^:]+)\\:([^:]+)$")

	for _, v := range parts {
		match := re.FindStringSubmatch(v)
		if match == nil || len(match) == 0 {
			sq.SearchTerm = v
			continue
		}

		switch qName, qVal := match[1], match[2]; qName {
		case "in":
			sq.In = qVal
		case "created":
			sq.Created = qVal
		case "updated":
			sq.Updated = qVal
		case "downloads":
			sq.Downloads = qVal
		case "owner":
			sq.Owner = qVal
		}
	}

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		page = 1
	}

	perPage, err := strconv.ParseUint(perPageStr, 10, 64)
	if err != nil {
		perPage = uint64(constants.PackagesQueryMaxPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	pkgRows, err := pkg.Search(&sq, &pc, &sc)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	pkgs := make([]*types.Package, len(pkgRows))

	for i, pr := range pkgRows {
		pkgs[i] = &types.Package{
			Name:             pr.Name,
			ID:               pr.ID,
			Desc:             pr.Description,
			Owner:            pr.OwnerUsername,
			Version:          pr.LatestVersion,
			Downloads:        pr.Downloads,
			PublishedAt:      pr.PublishedAt,
			UpdatedAt:        pr.LastReleasedAt,
			License:          pr.License,
			Homepage:         pr.HomepageURL,
			RepositoryURL:    pr.RepositoryURL,
			DocumentationURL: pr.DocumentationURL,
			BugsURL:          pr.BugsURL,
			Engines: types.Engines{
				Go: pr.EnginesGO,
			},
			Os: listOsNames(pr.OS),
		}
	}

	helper.WriteResponseJSONOk(w, r, pkgs)
}

// DownloadsGET returns the list of download counts for all public packages.
// Request: GET /downloads?page=1&per_page=10&sort=downloads&order=asc
// Only total downloads through registry: GET /downloads?onlyRegistry=true
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
// 5. id
func DownloadsGET(w http.ResponseWriter, r *http.Request) {
	qParams := r.URL.Query()

	var (
		onlyRegistry = qParams.Get("onlyRegistry")
		pageStr      = qParams.Get("page")
		perPageStr   = qParams.Get("per_page")
		sort         = qParams.Get("sort")
		order        = qParams.Get("order")
	)

	onlyRegistry = strings.ToLower(strings.TrimSpace(onlyRegistry))
	if onlyRegistry == "1" || onlyRegistry == "true" {
		rDownloads, err := pkg.RegistryTotalDownloads()
		if err != nil {
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		}

		rd := &types.RegistryDownloads{
			Downloads: rDownloads,
		}

		helper.WriteResponseJSONOk(w, r, rd)
		return
	}

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		page = 1
	}

	perPage, err := strconv.ParseUint(perPageStr, 10, 64)
	if err != nil {
		perPage = uint64(constants.PackagesQueryMaxPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	pkgRows, err := pkg.Search(nil, &pc, &sc)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	pDownloads := make([]*types.PackageDownloads, len(pkgRows))

	for i, pr := range pkgRows {
		pDownloads[i] = &types.PackageDownloads{
			Name:      pr.Name,
			ID:        pr.ID,
			Downloads: pr.Downloads,
		}
	}

	helper.WriteResponseJSONOk(w, r, pDownloads)
}

// SinglePackageVersionsGET returns the version histories of a single package.
// Request: GET /versions/:packageName
func SinglePackageVersionsGET(w http.ResponseWriter, r *http.Request) {
	inputPkgName := mux.Vars(r)["packageName"]

	vHistory, err := pkg.Versions(inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if vHistory == nil {
		errorCtrl.Error404(w, r)
		return
	}

	pvh := &types.PackageVersionHistory{
		Name:     vHistory.Name,
		ID:       vHistory.ID,
		Versions: make([]types.PackageVersion, len(*vHistory.Versions)),
	}

	for i, v := range *vHistory.Versions {
		pvh.Versions[i] = types.PackageVersion{
			Version:    v.Version,
			ReleasedAt: v.ReleasedAT,
		}
	}

	helper.WriteResponseJSONOk(w, r, pvh)
}

// SinglePackageReadmeGET returns the content of README file.
// Request: GET /packages/:packageName/readme
// For a specific version: GET /packages/:packageName/readme?v=1.0.2
func SinglePackageReadmeGET(w http.ResponseWriter, r *http.Request) {
	inputPkgName := mux.Vars(r)["packageName"]
	whereClause := "name = ?"
	sortBy := "id ASC"
	limit := "1"

	pkgRows, err := pkg.Query(whereClause, sortBy, limit, "", inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if len(pkgRows) == 0 {
		errorCtrl.Error404(w, r)
		return
	}

	inputVersion := strings.TrimSpace(r.URL.Query().Get("v"))

	if str.IsEmpty(inputVersion) {
		inputVersion = pkgRows[0].LatestVersion
	} else {
		inputVersion, err = helper.SanitizePackageVersion(inputVersion)
		if err != nil {
			errorCtrl.Error(w, r, http.StatusBadRequest, err.Error())
			return
		}

		ok, err := pkg.VersionExists(inputPkgName, inputVersion)
		if err != nil {
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		}

		if !ok {
			errorCtrl.Error(w, r, http.StatusNotFound, fmt.Sprintf("Version %s does not exist", inputVersion))
			return
		}
	}

	readmeData, err := vcs.PackageReadme(inputPkgName, inputVersion)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if readmeData == nil {
		errorCtrl.Error(w, r, http.StatusNotFound, fmt.Sprintf("No README found for the version %s", inputVersion))
		return
	}

	readmeResp := &types.PackageReadme{
		Name:    readmeData.Name,
		Version: readmeData.Version,
		Size:    readmeData.Size,
		Content: readmeData.Content,
	}

	helper.WriteResponseJSONOk(w, r, readmeResp)
}
