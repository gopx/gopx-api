package pkg

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/helper"
	"gopx.io/gopx-api/api/v1/controller/user"
	"gopx.io/gopx-api/api/v1/types"
	"gopx.io/gopx-api/pkg/controller/database"
	"gopx.io/gopx-api/pkg/controller/vcs"
	"gopx.io/gopx-common/arr"
	"gopx.io/gopx-common/fs"
	"gopx.io/gopx-common/misc"
	"gopx.io/gopx-common/str"
)

// SearchQuery holds the query values while searching the packages.
// Search pattern: gopx in:name downloads:>=1000 created:2017-01-01..2017-12-31
// Possible values of 'in' qualifier:
//	1. name
//	2. desc
//	3. tag
//	4. or any combination of them with comma separation.
// Note: Replace a whitespace with '+' character in query values.
type SearchQuery struct {
	SearchTerm string
	In         string
	Created    string
	Updated    string
	Downloads  string
	Owner      string
}

// QueryRow represents a single row to query a package data from database.
type QueryRow struct {
	ID               uint64
	Name             string
	OwnerUsername    string
	Downloads        uint64
	LatestVersion    string
	PublishedAt      time.Time
	LastReleasedAt   time.Time
	Description      string
	License          string
	HomepageURL      string
	RepositoryURL    string
	DocumentationURL string
	BugsURL          string
	EnginesGO        string
	OS               string
}

// VersionHistory represents the version histories for a single package.
type VersionHistory struct {
	Name     string
	ID       uint64
	Versions *[]*SingleVersion
}

// SingleVersion holds info of a single version.
type SingleVersion struct {
	Version    string
	ReleasedAT time.Time
}

// ReadmeData holds the package README content in base64 format.
type ReadmeData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    uint64 `json:"size"`
	Content string `json:"content"`
}

// Search searches packages according to the search query and returns a slice containing
// the results.
func Search(q *SearchQuery, pc *helper.PaginationConfig, sc *helper.SortingConfig) (pkgs []*QueryRow, err error) {
	if q == nil {
		q = &SearchQuery{}
	}
	if pc == nil {
		pc = &helper.PaginationConfig{}
	}
	if sc == nil {
		sc = &helper.SortingConfig{}
	}

	q.SearchTerm = helper.DecodeQueryValue(q.SearchTerm)
	q.In = helper.DecodeQueryValue(q.In)
	q.Created = helper.DecodeQueryValue(q.Created)
	q.Updated = helper.DecodeQueryValue(q.Updated)
	q.Downloads = helper.DecodeQueryValue(q.Downloads)
	q.Owner = helper.DecodeQueryValue(q.Owner)

	if str.IsEmpty(q.In) {
		q.In = strings.Join(constants.PackageQueryIns, ",")
	}

	whereClauses := []string{}
	placeholderValues := []interface{}{}

	// Add filters for q.SearchTerm and q.In
	if !str.IsEmpty(q.SearchTerm) {
		inClause := helper.PrepareInQualifierClause(&placeholderValues, constants.PackageQueryIns, constants.PackageQueryInsDbMap, q.SearchTerm, q.In)
		if !str.IsEmpty(inClause) {
			whereClauses = append(whereClauses, inClause)
		}
	}

	// Add filters for q.Created
	if !str.IsEmpty(q.Created) {
		cClause, err := helper.PrepareRelationalQueryClause(&placeholderValues, "published_at", q.Created)
		if err != nil {
			err = errors.Wrap(err, "Failed to create clause for q.Created")
			return nil, err
		}

		if !str.IsEmpty(cClause) {
			whereClauses = append(whereClauses, cClause)
		}
	}

	// Add filters for q.Updated
	if !str.IsEmpty(q.Updated) {
		uClause, err := helper.PrepareRelationalQueryClause(&placeholderValues, "last_released_at", q.Updated)
		if err != nil {
			err = errors.Wrap(err, "Failed to create clause for q.Updated")
			return nil, err
		}

		if !str.IsEmpty(uClause) {
			whereClauses = append(whereClauses, uClause)
		}
	}

	// Add filters for q.Downloads
	if !str.IsEmpty(q.Downloads) {
		dClause, err := helper.PrepareRelationalQueryClause(&placeholderValues, "downloads", q.Downloads)
		if err != nil {
			err = errors.Wrap(err, "Failed to create clause for q.Downloads")
			return nil, err
		}

		if !str.IsEmpty(dClause) {
			whereClauses = append(whereClauses, dClause)
		}
	}

	if !str.IsEmpty(q.Owner) {
		whereClauses = append(whereClauses, "owner_username = ?")
		placeholderValues = append(placeholderValues, q.Owner)
	}

	sanSortByCols := helper.SanitizeSortByCols(sc.SortBy, constants.PackageSortByCols)
	if len(sanSortByCols) == 0 {
		sanSortByCols = []string{constants.PackageDefaultSortByCol}
	}
	for i, v := range sanSortByCols {
		sanSortByCols[i] = constants.PackageSortByDbColsMap[arr.FindStr(constants.PackageSortByCols, v)]
	}

	sc.Order = strings.ToUpper(sc.Order)
	if str.IsEmpty(sc.Order) || arr.FindStr(constants.SortOrders, sc.Order) == -1 {
		sc.Order = constants.DefaultSortOrder
	}

	if pc.Page <= 0 {
		pc.Page = 1
	}
	if pc.PerPageCount <= 0 || pc.PerPageCount > uint64(constants.PackagesQueryMaxPageSize) {
		pc.PerPageCount = uint64(constants.PackagesQueryMaxPageSize)
	}
	qLimit, qOffset := pc.PerPageCount, (pc.Page-1)*pc.PerPageCount

	whereClauseSt := strings.Join(whereClauses, " and ")
	sortBySt := fmt.Sprintf("%s %s", strings.Join(sanSortByCols, ","), sc.Order)
	limitSt := fmt.Sprintf("%d", qLimit)
	offsetSt := fmt.Sprintf("%d", qOffset)

	pkgs, err = Query(whereClauseSt, sortBySt, limitSt, offsetSt, placeholderValues...)
	if err != nil {
		err = errors.Wrap(err, "Failed to query packages data from database")
		return nil, err
	}

	return pkgs, nil
}

// Query is a low-level function which queries package data from database
// according to the input filters.
func Query(whereClause, sortBy, limit, offset string, args ...interface{}) (pkgRows []*QueryRow, err error) {
	sqlSt := `
	SELECT DISTINCT
	packages.id, packages.name, packages.owner_username, packages.downloads, packages.latest_version, packages.published_at, packages.last_released_at,
	packages.description, packages.license, packages.homepage_url, packages.repository_url, packages.documentation_url, packages.bugs_url, packages.engines_go, packages.os
	FROM
	(SELECT
	packages.*, package_tags.tag
	FROM
	(SELECT
	packages.*, users.username AS owner_username
	FROM
	users
	INNER JOIN
	(SELECT packages.*, COUNT(package_downloads.id) AS downloads FROM packages LEFT JOIN package_downloads ON packages.id = package_downloads.package_id GROUP BY packages.id) AS packages
	ON
	users.id = packages.owner_id ORDER BY packages.id) AS packages
	LEFT JOIN
	package_tags
	ON
	packages.id = package_tags.package_id) AS packages
	`

	if !str.IsEmpty(whereClause) {
		sqlSt = fmt.Sprintf("%s WHERE %s", sqlSt, whereClause)
	}

	if !str.IsEmpty(sortBy) {
		sqlSt = fmt.Sprintf("%s ORDER BY %s", sqlSt, sortBy)
	}

	if !str.IsEmpty(limit) {
		sqlSt = fmt.Sprintf("%s LIMIT %s", sqlSt, limit)
	}

	if !str.IsEmpty(offset) {
		sqlSt = fmt.Sprintf("%s OFFSET %s", sqlSt, offset)
	}

	dbConn := database.Conn()

	rows, err := dbConn.Query(sqlSt, args...)
	if err != nil {
		err = errors.Wrap(err, "Failed to execute query statement")
		return nil, err
	}
	defer rows.Close()

	// Use minimum capacity of 10 for better performance
	pkgRows = make([]*QueryRow, 0, 10)

	var (
		id               uint64
		name             string
		ownerUsername    string
		downloads        uint64
		latestVersion    string
		publishedAt      time.Time
		lastReleasedAt   time.Time
		description      sql.RawBytes
		license          sql.RawBytes
		homepageURL      sql.RawBytes
		repositoryURL    sql.RawBytes
		documentationURL sql.RawBytes
		bugsURL          sql.RawBytes
		enginesGO        sql.RawBytes
		os               sql.RawBytes
	)

	for rows.Next() {
		err := rows.Scan(
			&id,
			&name,
			&ownerUsername,
			&downloads,
			&latestVersion,
			&publishedAt,
			&lastReleasedAt,
			&description,
			&license,
			&homepageURL,
			&repositoryURL,
			&documentationURL,
			&bugsURL,
			&enginesGO,
			&os,
		)
		if err != nil {
			err = errors.Wrap(err, "Failed to scan the package query result")
			return nil, err
		}

		qr := &QueryRow{
			ID:               id,
			Name:             name,
			OwnerUsername:    ownerUsername,
			Downloads:        downloads,
			LatestVersion:    latestVersion,
			PublishedAt:      publishedAt,
			LastReleasedAt:   lastReleasedAt,
			Description:      string(description),
			License:          string(license),
			HomepageURL:      string(homepageURL),
			RepositoryURL:    string(repositoryURL),
			DocumentationURL: string(documentationURL),
			BugsURL:          string(bugsURL),
			EnginesGO:        string(enginesGO),
			OS:               string(os),
		}

		pkgRows = append(pkgRows, qr)
	}

	if err := rows.Err(); err != nil {
		err = errors.Wrap(err, "Failed to fetch the package query result")
		return nil, err
	}

	return pkgRows, nil
}

// RegistryTotalDownloads calculates the total package downloads across registry.
func RegistryTotalDownloads() (downloads uint64, err error) {
	sqlSt := `
	SELECT
	COUNT(*) AS downloads
	FROM 
	package_downloads
	`

	dbConn := database.Conn()

	var totDownloads uint64
	err = dbConn.QueryRow(sqlSt).Scan(&totDownloads)
	if err != nil {
		err = errors.Wrap(err, "Failed to execute query statement")
		return
	}
	downloads = totDownloads

	return
}

// Versions returns the version histories of a single package.
func Versions(packageName string) (vHistory *VersionHistory, err error) {
	sqlSt := `
	SELECT
	*
	FROM
	(SELECT
	package_versions.id, package_versions.version, package_versions.released_at, packages.id AS package_id, packages.name AS package_name 
	FROM
	packages
	INNER JOIN
	package_versions
	ON
	packages.id = package_versions.package_id) as package_versions
	`

	sqlSt = fmt.Sprintf("%s WHERE package_name = ? ORDER BY released_at ASC", sqlSt)

	dbConn := database.Conn()

	rows, err := dbConn.Query(sqlSt, packageName)
	if err != nil {
		err = errors.Wrap(err, "Failed to execute query statement")
		return nil, err
	}
	defer rows.Close()

	vHistory = &VersionHistory{}
	versions := []*SingleVersion{}

	var (
		id            uint64
		version       string
		releasedAt    time.Time
		packageID     uint64
		packageNameDb string
	)

	for rows.Next() {
		err := rows.Scan(
			&id,
			&version,
			&releasedAt,
			&packageID,
			&packageNameDb,
		)
		if err != nil {
			err = errors.Wrap(err, "Failed to scan the package version query result")
			return nil, err
		}

		vHistory.ID = packageID
		vHistory.Name = packageNameDb

		sv := &SingleVersion{
			Version:    version,
			ReleasedAT: releasedAt,
		}
		versions = append(versions, sv)
	}

	if err := rows.Err(); err != nil {
		err = errors.Wrap(err, "Failed to fetch the package version query result")
		return nil, err
	}

	vHistory.Versions = &versions

	if vHistory.ID == 0 || str.IsEmpty(vHistory.Name) {
		vHistory = nil
	}

	return vHistory, nil
}

// DeletePackage deletes a specific package by removing its data from
// database and from vcs registry.
func DeletePackage(packageName string) (err error) {
	sqlSt := `
	SELECT id
	FROM packages
	WHERE name = ? ORDER BY id ASC LIMIT 1
	`

	dbConn := database.Conn()

	var packageID uint64
	err = dbConn.QueryRow(sqlSt, packageName).Scan(&packageID)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			err = errors.Wrap(err, "Package not found")
		default:
			err = errors.Wrap(err, "Failed to query package id")
		}
		return
	}

	tx, err := dbConn.Begin()
	if err != nil {
		err = errors.Wrap(err, "Failed to begin a transaction")
		return
	}

	st := `
	DELETE FROM packages
	WHERE id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from packages table")
		return
	}

	st = `
	DELETE FROM package_versions
	WHERE package_id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from package_versions table")
		return
	}

	st = `
	DELETE FROM package_tags
	WHERE package_id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from package_tags table")
		return
	}

	st = `
	DELETE FROM package_readme
	WHERE package_id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from package_readme table")
		return
	}

	st = `
	DELETE FROM package_downloads
	WHERE package_id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from package_downloads table")
		return
	}

	err = vcs.DeletePackage(packageName)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete package data from vcs registry")
		return
	}

	err = tx.Commit()
	// Possibility is very low.
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to commit changes of deleting package")
		return
	}

	return nil
}

// SanitizePackageMeta sanitizes the input metadata.
func SanitizePackageMeta(meta *types.PackageMetaData) (err error) {
	meta.Name = strings.TrimSpace(meta.Name)
	err = helper.ValidatePackageName(meta.Name)
	if err != nil {
		return
	}

	meta.Version = strings.TrimSpace(meta.Version)
	meta.Version, err = helper.SanitizePackageVersion(meta.Version)
	if err != nil {
		return
	}

	if meta.Tags == nil {
		meta.Tags = []string{}
	}
	sTags := []string{}
	for _, v := range meta.Tags {
		v = strings.ToLower(strings.TrimSpace(v))
		if !str.IsEmpty(v) {
			sTags = append(sTags, v)
		}
	}
	meta.Tags = sTags

	if meta.Os == nil {
		meta.Os = []string{}
	}
	sOs := []string{}
	for _, v := range meta.Os {
		v = strings.TrimSpace(v)
		parts := strings.Split(v, ":")

		v = strings.TrimSpace(parts[0])
		if len(parts) >= 2 {
			parts[1] = strings.TrimSpace(parts[1])
			if !str.IsEmpty(parts[1]) {
				v += ":" + parts[1]
			}
		}

		if !str.IsEmpty(v) {
			sOs = append(sOs, v)
		}
	}
	meta.Os = sOs

	return
}

// InsertNew inserts a new package to the database and registers to the vcs registry.
func InsertNew(meta *types.PackageMetaData, data io.ReadSeeker, ownerInfo *user.QueryRow) (pkg *QueryRow, err error) {
	dbConn := database.Conn()
	tx, err := dbConn.Begin()
	if err != nil {
		err = errors.Wrap(err, "Failed to begin a transaction")
		return
	}

	st := `
	INSERT INTO packages
	(name, owner_id, latest_version, description, license, homepage_url, repository_url, documentation_url, bugs_url, engines_go, os)
	VALUES 
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	r, err := tx.Exec(
		st,
		meta.Name,
		ownerInfo.ID,
		meta.Version,
		meta.Description,
		meta.License,
		meta.HomepageURL,
		meta.RepositoryURL,
		meta.DocumentationURL,
		meta.BugsURL,
		meta.Engines.Go,
		strings.Join(meta.Os, ", "),
	)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to insert package data to packages table")
		return
	}

	packageID, err := r.LastInsertId()
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to retrive the last insert ID")
		return
	}

	st = `
	INSERT INTO package_versions
	(version, package_id)
	VALUES 
	(?, ?)
	`
	_, err = tx.Exec(st, meta.Version, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to insert package data to package_versions table")
		return
	}

	st = `
	INSERT INTO package_tags
	(package_id, tag)
	VALUES 
	(?, ?)
	`
	prepSt, err := tx.Prepare(st)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to create prepared statement")
		return
	}

	for _, v := range meta.Tags {
		_, err = prepSt.Exec(packageID, v)
		if err != nil {
			tx.Rollback()
			err = errors.Wrap(err, "Failed to insert tags to package_tags table")
			return
		}
	}
	prepSt.Close()

	var readmeBuff bytes.Buffer
	ok, idx, err := fs.ReadEntryTarGz(data, &readmeBuff, constants.ReadmeFileNames)
	if err != nil {
		tx.Rollback()
		err = errors.Wrapf(err, "Failed to retrive README file")
		return
	}

	var readmeFileName string
	var readmeContent []byte
	if ok {
		readmeFileName = constants.ReadmeFileNames[idx]
		readmeContent = readmeBuff.Bytes()
		if len(readmeContent) == 0 {
			readmeContent = defaultPackageReadme(meta.Name)
		}
	} else {
		readmeFileName = constants.DefaultReadmeFileName
		readmeContent = defaultPackageReadme(meta.Name)
	}

	st = `
	INSERT INTO package_readme
	(package_id, version, name, file_size, content)
	VALUES 
	(?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(st, packageID, meta.Version, readmeFileName, len(readmeContent), readmeContent)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to insert package data to package_readme table")
		return
	}

	_, err = data.Seek(0, 0)
	if err != nil {
		tx.Rollback()
		err = errors.Wrapf(err, "Failed to seek the data reader to starting position")
		return
	}

	vcsMeta := &vcs.PackageMeta{
		Type:    vcs.PackageTypePublic,
		Name:    meta.Name,
		Version: meta.Version,
		Owner: vcs.PackageOwner{
			Name:        ownerInfo.Name,
			PublicEmail: misc.TerOpt(ownerInfo.IsPublicEmail, ownerInfo.Email, "").(string),
			Username:    ownerInfo.Username,
		},
	}
	err = vcs.RegisterPackage(vcsMeta, data)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to register package to vcs registry")
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to commit changes")
		return
	}

	pkgRows, err := Query("id = ?", "id ASC", "1", "", packageID)
	if err != nil || len(pkgRows) == 0 {
		err = errors.Wrap(err, "Failed to read updated package data")
		return
	}

	return pkgRows[0], nil
}

// MakeNewRelease creates a new release/version to the database and registers that version to the vcs registry.
func MakeNewRelease(packageID uint64, meta *types.PackageMetaData, data io.ReadSeeker, ownerInfo *user.QueryRow) (pkg *QueryRow, err error) {
	dbConn := database.Conn()
	tx, err := dbConn.Begin()
	if err != nil {
		err = errors.Wrap(err, "Failed to begin a transaction")
		return
	}

	st := `
	UPDATE packages
	SET latest_version = ?, description = ?, license = ?, homepage_url = ?, repository_url = ?, documentation_url = ?, bugs_url = ?, engines_go = ?, os = ?
	WHERE id = ?
	`
	_, err = tx.Exec(
		st,
		meta.Version,
		meta.Description,
		meta.License,
		meta.HomepageURL,
		meta.RepositoryURL,
		meta.DocumentationURL,
		meta.BugsURL,
		meta.Engines.Go,
		strings.Join(meta.Os, ", "),
		packageID,
	)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to update package data to packages table")
		return
	}

	st = `
	INSERT INTO package_versions
	(version, package_id)
	VALUES
	(?, ?)
	`
	_, err = tx.Exec(st, meta.Version, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to insert package data to package_versions table")
		return
	}

	st = `
	DELETE FROM package_tags
	WHERE package_id = ?
	`
	_, err = tx.Exec(st, packageID)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to delete existing tags from package_tags table")
		return
	}

	st = `
	INSERT INTO package_tags
	(package_id, tag)
	VALUES 
	(?, ?)
	`
	prepSt, err := tx.Prepare(st)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to create prepared statement")
		return
	}

	for _, v := range meta.Tags {
		_, err = prepSt.Exec(packageID, v)
		if err != nil {
			tx.Rollback()
			err = errors.Wrap(err, "Failed to insert tags to package_tags table")
			return
		}
	}
	prepSt.Close()

	var readmeBuff bytes.Buffer
	ok, idx, err := fs.ReadEntryTarGz(data, &readmeBuff, constants.ReadmeFileNames)
	if err != nil {
		tx.Rollback()
		err = errors.Wrapf(err, "Failed to retrive README file")
		return
	}

	var readmeFileName string
	var readmeContent []byte
	if ok {
		readmeFileName = constants.ReadmeFileNames[idx]
		readmeContent = readmeBuff.Bytes()
		if len(readmeContent) == 0 {
			readmeContent = defaultPackageReadme(meta.Name)
		}
	} else {
		readmeFileName = constants.DefaultReadmeFileName
		readmeContent = defaultPackageReadme(meta.Name)
	}

	st = `
	INSERT INTO package_readme
	(package_id, version, name, file_size, content)
	VALUES 
	(?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(st, packageID, meta.Version, readmeFileName, len(readmeContent), readmeContent)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to insert package data to package_readme table")
		return
	}

	_, err = data.Seek(0, 0)
	if err != nil {
		tx.Rollback()
		err = errors.Wrapf(err, "Failed to seek the data reader to starting position")
		return
	}

	vcsMeta := &vcs.PackageMeta{
		Type:    vcs.PackageTypePublic,
		Name:    meta.Name,
		Version: meta.Version,
		Owner: vcs.PackageOwner{
			Name:        ownerInfo.Name,
			PublicEmail: misc.TerOpt(ownerInfo.IsPublicEmail, ownerInfo.Email, "").(string),
			Username:    ownerInfo.Username,
		},
	}
	err = vcs.RegisterPackage(vcsMeta, data)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to register package's new version to vcs registry")
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to commit changes")
		return
	}

	pkgRows, err := Query("id = ?", "id ASC", "1", "", packageID)
	if err != nil || len(pkgRows) == 0 {
		err = errors.Wrap(err, "Failed to read updated package data")
		return
	}

	return pkgRows[0], nil
}

// VersionExists checks whether the input version exists or not.
func VersionExists(pkgName, inpVersion string) (ok bool, err error) {
	vh, err := Versions(pkgName)
	if err != nil {
		return
	}

	for _, v := range *vh.Versions {
		if ok, err := helper.IsSameVersion(v.Version, inpVersion); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}

	return
}

func defaultPackageReadme(pkgName string) (content []byte) {
	content = []byte(fmt.Sprintf("No readme found for package %s", pkgName))
	return
}

// Readme returns the README content of a package.
func Readme(packageID uint64, version string) (content *ReadmeData, err error) {
	sqlSt := `
	SELECT name, file_size, content
	FROM package_readme
	WHERE package_id = ? and version = ?
	ORDER BY id ASC
	LIMIT 1
	`
	var (
		readmeName    string
		size          uint64
		readmeContent []byte
	)

	dbConn := database.Conn()
	err = dbConn.QueryRow(sqlSt, packageID, version).Scan(&readmeName, &size, &readmeContent)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			err = errors.Wrapf(err, "Package not found")
			return
		default:
			err = errors.Wrapf(err, "Failed to read README data from package_readme table")
			return
		}
	}

	content = &ReadmeData{
		Name:    readmeName,
		Version: version,
		Size:    size,
		Content: base64.StdEncoding.EncodeToString(readmeContent),
	}

	return
}
