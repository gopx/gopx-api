package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	yaml "gopkg.in/yaml.v2"
	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/helper"
	"gopx.io/gopx-api/api/v1/controller/pkg"
	"gopx.io/gopx-api/api/v1/controller/user"
	"gopx.io/gopx-api/api/v1/types"
	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
	"gopx.io/gopx-common/fs"
	"gopx.io/gopx-common/log"
	"gopx.io/gopx-common/misc"
	"gopx.io/gopx-common/str"
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
		perPage = uint64(constants.UsersQueryMaxPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	userRows, err := user.Search(nil, &pc, &sc)
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

	helper.WriteResponseValueOK(w, r, users)
}

// SingleUserGET returns the public information about a single user.
// Request: GET /users/:username
func SingleUserGET(w http.ResponseWriter, r *http.Request) {
	inputUsername := mux.Vars(r)["username"]
	whereClause := "username = ?"
	sortBy := "id ASC"
	limit := "1"

	userRows, err := user.Query(whereClause, sortBy, limit, "", inputUsername)
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

	helper.WriteResponseValueOK(w, r, user)
}

// CurrentUserGET returns the public and private information of the authenticated user.
// Request: GET /user
func CurrentUserGET(w http.ResponseWriter, r *http.Request) {
	ur, err := authUser(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if ur == nil {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	user := &types.User{
		Username:      ur.Username,
		ID:            ur.ID,
		Name:          ur.Name,
		Email:         ur.Email,
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

	helper.WriteResponseValueOK(w, r, user)
}

// SearchUsersGET performs a search query among all users.
// Request: GET /search/users?q=rousan+in:name,username+joined:2017-01-01..2017-12-31&sort=packages&order=desc&page=1&per_page=10
// Sorting can be performed on:
// 1. packages
// 2. joined
// 4. username
// 5. id
func SearchUsersGET(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	var (
		q          = params.Get("q")
		pageStr    = params.Get("page")
		perPageStr = params.Get("per_page")
		sort       = params.Get("sort")
		order      = params.Get("order")
	)

	sq := user.SearchQuery{}
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
		case "joined":
			sq.Joined = qVal
		case "location":
			sq.Location = qVal
		case "packages":
			sq.Packages = qVal
		}
	}

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		page = 1
	}

	perPage, err := strconv.ParseUint(perPageStr, 10, 64)
	if err != nil {
		perPage = uint64(constants.UsersQueryMaxPageSize)
	}

	pc := helper.PaginationConfig{
		Page:         page,
		PerPageCount: perPage,
	}

	sc := helper.SortingConfig{
		SortBy: sort,
		Order:  order,
	}

	userRows, err := user.Search(&sq, &pc, &sc)
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

	helper.WriteResponseValueOK(w, r, users)
}

// CurrentUserPATCH updates data to the authenticated user.
// Request: PATCH /user
func CurrentUserPATCH(w http.ResponseWriter, r *http.Request) {
	ur, err := authUser(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if ur == nil {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	inputData := types.UserMutation{}
	err = json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Problems parsing JSON data")
		return
	}

	if inputData.Social == nil {
		inputData.Social = &types.UserSocialAccountsMutation{}
	}

	updateData := &user.MutationData{
		Name:         inputData.Name,
		URL:          inputData.Blog,
		Organization: inputData.Organization,
		Location:     inputData.Location,
		Social: &user.SocialAccountsMutationData{
			Github:        inputData.Social.Github,
			Twitter:       inputData.Social.Twitter,
			StackOverflow: inputData.Social.StackOverflow,
			LinkedIn:      inputData.Social.LinkedIn,
		},
	}

	ur, err = user.UpdateInfo(ur.ID, updateData)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	user := &types.User{
		Username:      ur.Username,
		ID:            ur.ID,
		Name:          ur.Name,
		Email:         ur.Email,
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

	helper.WriteResponseValueOK(w, r, user)
}

// SingleUserPackagesGET returns the list of public packages of a single user.
// Request: GET /users/:username/packages?sort=downloads,id&order=desc&page=1&per_page=10
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
// 5. id
func SingleUserPackagesGET(w http.ResponseWriter, r *http.Request) {
	inputUsername := mux.Vars(r)["username"]

	userRows, err := user.Query("username = ?", "id ASC", "1", "", inputUsername)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if len(userRows) == 0 {
		errorCtrl.Error404(w, r)
		return
	}

	params := r.URL.Query()

	var (
		pageStr    = params.Get("page")
		perPageStr = params.Get("per_page")
		sort       = params.Get("sort")
		order      = params.Get("order")
	)

	sq := pkg.SearchQuery{
		Owner: inputUsername,
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

	helper.WriteResponseValueOK(w, r, pkgs)
}

// CurrentUserPackagesGET returns the list of all packages of the authenticated user.
// Request: GET /user/packages?sort=downloads,id&order=desc&page=1&per_page=10
// Sorting can be performed on:
// 1. downloads
// 2. created
// 3. updated
// 4. name
// 5. id
func CurrentUserPackagesGET(w http.ResponseWriter, r *http.Request) {
	ur, err := authUser(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if ur == nil {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	params := r.URL.Query()

	var (
		pageStr    = params.Get("page")
		perPageStr = params.Get("per_page")
		sort       = params.Get("sort")
		order      = params.Get("order")
	)

	sq := pkg.SearchQuery{
		Owner: ur.Username,
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

	helper.WriteResponseValueOK(w, r, pkgs)
}

// CurrentUserPackagesPOST register a new package of a authenticated user.
// Request: POST /user/packages
func CurrentUserPackagesPOST(w http.ResponseWriter, r *http.Request) {
	ur, err := authUser(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if ur == nil {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Content-Type must be multipart/form-data")
		return
	}

	tmpFile, err := ioutil.TempFile("", constants.TempFileNamePrefixForPackageRegister)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}
	defer os.RemoveAll(tmpFile.Name())
	defer tmpFile.Close()

	ok, err := readPackageData(mr, tmpFile)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if !ok {
		errorCtrl.Error(w, r, http.StatusBadRequest, "Package data not found with param name data")
		return
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	var wBuff bytes.Buffer
	ok, idx, err := fs.ReadEntryTarGz(tmpFile, &wBuff, constants.PackageMetaFileNames)
	if err != nil {
		errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid data or exceeds maximum allowed size %dMB", constants.PackageDataMaxSize/(1024*1024)))
		return
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if !ok {
		errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("The meta file %s not found in package contents", strings.Join(constants.PackageMetaFileNames, " or ")))
		return
	}

	metaFileName := constants.PackageMetaFileNames[idx]
	meta := types.PackageMetaData{}
	if metaFileName == "gopx.json" {
		err = json.NewDecoder(&wBuff).Decode(&meta)
		if err != nil {
			errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Problems parsing %s file", metaFileName))
			return
		}
	} else {
		err = yaml.NewDecoder(&wBuff).Decode(&meta)
		if err != nil {
			errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Problems parsing %s file", metaFileName))
			return
		}
	}

	err = pkg.SanitizePackageMeta(&meta)
	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	pkgRows, err := pkg.Query("name = ?", "id ASC", "1", "", meta.Name)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	var fPkg *pkg.QueryRow

	if len(pkgRows) < 1 {
		iPkg, err := pkg.InsertNew(&meta, tmpFile, ur)
		if err != nil {
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		}
		fPkg = iPkg
	} else {
		pkgRow := pkgRows[0]

		if pkgRow.OwnerUsername != ur.Username {
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
			return
		}

		ok, err := helper.IsSameVersion(pkgRow.LatestVersion, meta.Version)
		if err != nil {
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		}

		if ok {
			errorCtrl.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Package version %s already exists", meta.Version))
			return
		}

		uPkg, err := pkg.MakeNewRelease(pkgRow.ID, &meta, tmpFile, ur)
		if err != nil {
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		}

		fPkg = uPkg
	}

	pkgData := &types.Package{
		Name:             fPkg.Name,
		ID:               fPkg.ID,
		Desc:             fPkg.Description,
		Owner:            fPkg.OwnerUsername,
		Version:          fPkg.LatestVersion,
		Downloads:        fPkg.Downloads,
		PublishedAt:      fPkg.PublishedAt,
		UpdatedAt:        fPkg.LastReleasedAt,
		License:          fPkg.License,
		Homepage:         fPkg.HomepageURL,
		RepositoryURL:    fPkg.RepositoryURL,
		DocumentationURL: fPkg.DocumentationURL,
		BugsURL:          fPkg.BugsURL,
		Engines: types.Engines{
			Go: fPkg.EnginesGO,
		},
		Os: listOsNames(fPkg.OS),
	}

	helper.WriteResponseValue(w, r, pkgData, http.StatusCreated)
}

func readPackageData(mr *multipart.Reader, w io.Writer) (ok bool, err error) {
	for {
		p, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return false, err
			}
		}

		ok, err := readSinglePart(p, w)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}

func readSinglePart(p *multipart.Part, w io.Writer) (ok bool, err error) {
	defer p.Close()
	if p.FormName() == constants.PackageUploadParamName {
		_, err = io.CopyN(w, p, constants.PackageDataMaxSize)
		if err == nil || err == io.EOF {
			ok = true
			err = nil
		}
	}

	return
}

// CurrentUserPackagesDELETE deletes a whole package and free up the package name.
// Request: DELETE /user/packages/:packageName
func CurrentUserPackagesDELETE(w http.ResponseWriter, r *http.Request) {
	inputPkgName := mux.Vars(r)["packageName"]

	pkgRows, err := pkg.Query("name = ?", "id ASC", "1", "", inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	if len(pkgRows) == 0 {
		errorCtrl.Error404(w, r)
		return
	}

	ur, err := authUser(r.Header.Get("Authorization"))

	if err != nil {
		switch err {
		case constants.ErrInternalServer:
			log.Error("Error %s", err)
			errorCtrl.Error500(w, r)
			return
		default:
			errorCtrl.Error(w, r, http.StatusUnauthorized, "Requires authentication")
			return
		}
	}

	if ur == nil {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	// Important
	if ur.Username != pkgRows[0].OwnerUsername {
		errorCtrl.Error(w, r, http.StatusUnauthorized, "Bad credentials")
		return
	}

	err = pkg.DeletePackage(inputPkgName)
	if err != nil {
		log.Error("Error %s", err)
		errorCtrl.Error500(w, r)
		return
	}

	helper.WriteResponse(w, r, nil, http.StatusNoContent)
}
