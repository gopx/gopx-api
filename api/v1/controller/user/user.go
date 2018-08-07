package user

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/helper"
	"gopx.io/gopx-api/pkg/controller/database"
	"gopx.io/gopx-common/arr"
	"gopx.io/gopx-common/misc"
	"gopx.io/gopx-common/str"
)

// SearchQuery holds the query values while searching the users.
// Search pattern: gopx in:username packages:<=100 location:india joined:2017-01-01..2017-12-31
// Possible values of 'in' qualifier:
// 	1. username
//	2. email
// 	3. name
// 	4. or any combination of them with comma separation.
// Note: Replace a whitespace with '+' character in query values.
type SearchQuery struct {
	SearchTerm string
	In         string
	Packages   string
	Location   string
	Joined     string
}

// QueryRow represents a single row to query an user data from database.
type QueryRow struct {
	ID            uint64
	Name          string
	PackagesCount uint64
	JoinedAt      time.Time
	Email         string
	IsPublicEmail bool
	Username      string
	Password      string
	Avatar        string
	URL           string
	Organization  string
	Location      string
	Github        string
	Twitter       string
	StackOverflow string
	LinkedIn      string
	APIKey        string
}

// MutationData holds user mutation data.
type MutationData struct {
	Name         *string
	URL          *string
	Organization *string
	Location     *string
	Social       *SocialAccountsMutationData
}

// SocialAccountsMutationData holds user social account mutation data.
type SocialAccountsMutationData struct {
	Github        *string `json:"github"`
	Twitter       *string `json:"twitter"`
	StackOverflow *string `json:"stackOverflow"`
	LinkedIn      *string `json:"linkedin"`
}

// Search searches users according to the search query and returns a slice containing
// the results.
func Search(q *SearchQuery, pc *helper.PaginationConfig, sc *helper.SortingConfig) (users []*QueryRow, err error) {
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
	q.Packages = helper.DecodeQueryValue(q.Packages)
	q.Location = helper.DecodeQueryValue(q.Location)
	q.Joined = helper.DecodeQueryValue(q.Joined)

	if str.IsEmpty(q.In) {
		q.In = strings.Join(constants.UserQueryIns, ",")
	}

	whereClauses := []string{}
	placeholderValues := []interface{}{}

	// Add filters for q.SearchTerm and q.In
	if !str.IsEmpty(q.SearchTerm) {
		inClause := helper.PrepareInQualifierClause(&placeholderValues, constants.UserQueryIns, constants.UserQueryInsDbMap, q.SearchTerm, q.In)
		if !str.IsEmpty(inClause) {
			whereClauses = append(whereClauses, inClause)
		}
	}

	// Add filters for q.Packages
	if !str.IsEmpty(q.Packages) {
		pClause, err := helper.PrepareRelationalQueryClause(&placeholderValues, "packages_count", q.Packages)
		if err != nil {
			err = errors.Wrap(err, "Failed to create clause for q.Packages")
			return nil, err
		}

		if !str.IsEmpty(pClause) {
			whereClauses = append(whereClauses, pClause)
		}
	}

	// Add filters for q.Location
	if !str.IsEmpty(q.Location) {
		lClauses := []string{}
		helper.PrepareMultiWordsQueryClause(&lClauses, &placeholderValues, "location", q.Location)

		if len(lClauses) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(lClauses, " or ")+")")
		}
	}

	// Add filters for q.Joined
	if !str.IsEmpty(q.Joined) {
		jClause, err := helper.PrepareRelationalQueryClause(&placeholderValues, "joined_at", q.Joined)
		if err != nil {
			err = errors.Wrap(err, "Failed to create clause for q.Joined")
			return nil, err
		}

		if !str.IsEmpty(jClause) {
			whereClauses = append(whereClauses, jClause)
		}
	}

	sanSortByCols := helper.SanitizeSortByCols(sc.SortBy, constants.UserSortByCols)
	if len(sanSortByCols) == 0 {
		sanSortByCols = []string{constants.UserDefaultSortByCol}
	}
	for i, v := range sanSortByCols {
		sanSortByCols[i] = constants.UserSortByDbColsMap[arr.FindStr(constants.UserSortByCols, v)]
	}

	sc.Order = strings.ToUpper(sc.Order)
	if str.IsEmpty(sc.Order) || arr.FindStr(constants.SortOrders, sc.Order) == -1 {
		sc.Order = constants.DefaultSortOrder
	}

	if pc.Page <= 0 {
		pc.Page = 1
	}
	if pc.PerPageCount <= 0 || pc.PerPageCount > uint64(constants.UsersQueryMaxPageSize) {
		pc.PerPageCount = uint64(constants.UsersQueryMaxPageSize)
	}
	qLimit, qOffset := pc.PerPageCount, (pc.Page-1)*pc.PerPageCount

	whereClauseSt := strings.Join(whereClauses, " and ")
	sortBySt := fmt.Sprintf("%s %s", strings.Join(sanSortByCols, ","), sc.Order)
	limitSt := fmt.Sprintf("%d", qLimit)
	offsetSt := fmt.Sprintf("%d", qOffset)

	users, err = Query(whereClauseSt, sortBySt, limitSt, offsetSt, placeholderValues...)
	if err != nil {
		err = errors.Wrap(err, "Failed to query users data from database")
		return nil, err
	}

	return users, nil
}

// Query is a low-level function which queries user data from database
// according to the input filters.
func Query(whereClause, sortBy, limit, offset string, args ...interface{}) (userRows []*QueryRow, err error) {
	sqlSt := `
	SELECT
	*
	FROM
  (SELECT
	users.*, user_api_keys.api_key
	FROM
	(SELECT
	users.id, users.name, users.packages_count, users.joined_at, users.email, users.is_public_email, users.username, users.password, users.avatar, users.url,
	users.organization, users.location, user_social_accounts.github, user_social_accounts.twitter, user_social_accounts.stack_overflow, user_social_accounts.linkedin
	FROM
	(SELECT users.*, COUNT(packages.id) AS packages_count FROM users LEFT JOIN packages ON users.id = packages.owner_id GROUP BY users.id) AS users
	INNER JOIN
	user_social_accounts
	ON
	users.id = user_social_accounts.user_id) AS users
	LEFT JOIN
	user_api_keys
	ON
	users.id = user_api_keys.user_id) AS users
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
	userRows = make([]*QueryRow, 0, 10)

	var (
		id            uint64
		name          string
		packagesCount uint64
		joinedAt      time.Time
		email         string
		isPublicEmail uint8
		username      string
		password      string
		avatar        string
		url           sql.RawBytes
		organization  sql.RawBytes
		location      sql.RawBytes
		github        sql.RawBytes
		twitter       sql.RawBytes
		stackOverflow sql.RawBytes
		linkedIn      sql.RawBytes
		apiKey        sql.RawBytes
	)
	for rows.Next() {
		err := rows.Scan(
			&id,
			&name,
			&packagesCount,
			&joinedAt,
			&email,
			&isPublicEmail,
			&username,
			&password,
			&avatar,
			&url,
			&organization,
			&location,
			&github,
			&twitter,
			&stackOverflow,
			&linkedIn,
			&apiKey,
		)
		if err != nil {
			err = errors.Wrap(err, "Failed to scan the user query result")
			return nil, err
		}

		qr := &QueryRow{
			ID:            id,
			Name:          name,
			PackagesCount: packagesCount,
			JoinedAt:      joinedAt,
			Email:         email,
			IsPublicEmail: misc.TerOpt(isPublicEmail == 1, true, false).(bool),
			Username:      username,
			Password:      password,
			Avatar:        avatar,
			URL:           string(url),
			Organization:  string(organization),
			Location:      string(location),
			Github:        string(github),
			Twitter:       string(twitter),
			StackOverflow: string(stackOverflow),
			LinkedIn:      string(linkedIn),
			APIKey:        string(apiKey),
		}

		userRows = append(userRows, qr)
	}

	if err := rows.Err(); err != nil {
		err = errors.Wrap(err, "Failed to fetch the user query result")
		return nil, err
	}

	return userRows, nil
}

// UpdateInfo updates user data and returns the updated one.
func UpdateInfo(userID uint64, data *MutationData) (user *QueryRow, err error) {
	sqlStUsers := []string{}
	placeholderValUsers := []interface{}{}

	if data.Name != nil {
		sqlStUsers = append(sqlStUsers, "name = ?")
		placeholderValUsers = append(placeholderValUsers, *data.Name)
	}

	if data.URL != nil {
		sqlStUsers = append(sqlStUsers, "url = ?")
		placeholderValUsers = append(placeholderValUsers, *data.URL)
	}

	if data.Organization != nil {
		sqlStUsers = append(sqlStUsers, "organization = ?")
		placeholderValUsers = append(placeholderValUsers, *data.Organization)
	}

	if data.Location != nil {
		sqlStUsers = append(sqlStUsers, "location = ?")
		placeholderValUsers = append(placeholderValUsers, *data.Location)
	}

	sqlStUserSocial := []string{}
	placeholderValUserSocial := []interface{}{}

	if data.Social.Github != nil {
		sqlStUserSocial = append(sqlStUserSocial, "github = ?")
		placeholderValUserSocial = append(placeholderValUserSocial, *data.Social.Github)
	}

	if data.Social.Twitter != nil {
		sqlStUserSocial = append(sqlStUserSocial, "twitter = ?")
		placeholderValUserSocial = append(placeholderValUserSocial, *data.Social.Twitter)
	}

	if data.Social.StackOverflow != nil {
		sqlStUserSocial = append(sqlStUserSocial, "stack_overflow = ?")
		placeholderValUserSocial = append(placeholderValUserSocial, *data.Social.StackOverflow)
	}

	if data.Social.LinkedIn != nil {
		sqlStUserSocial = append(sqlStUserSocial, "linkedin = ?")
		placeholderValUserSocial = append(placeholderValUserSocial, *data.Social.LinkedIn)
	}

	dbConn := database.Conn()
	tx, err := dbConn.Begin()
	if err != nil {
		err = errors.Wrap(err, "Failed to begin a transaction")
		return
	}

	if len(sqlStUsers) > 0 {
		st := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = ?`, strings.Join(sqlStUsers, ", "))
		placeholderValUsers = append(placeholderValUsers, userID)

		_, err = tx.Exec(st, placeholderValUsers...)
		if err != nil {
			tx.Rollback()
			err = errors.Wrap(err, "Failed to update users table")
			return
		}
	}

	if len(sqlStUserSocial) > 0 {
		st := fmt.Sprintf(`
		UPDATE user_social_accounts
		SET %s
		WHERE user_id = ?`, strings.Join(sqlStUserSocial, ", "))
		placeholderValUserSocial = append(placeholderValUserSocial, userID)

		_, err = tx.Exec(st, placeholderValUserSocial...)
		if err != nil {
			tx.Rollback()
			err = errors.Wrap(err, "Failed to update user_social_accounts table")
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "Failed to commit changes of user data")
		return
	}

	userRows, err := Query("id = ?", "id ASC", "1", "", userID)
	if err != nil || len(userRows) < 1 {
		err = errors.Wrap(err, "Failed to read updated user data")
		return
	}

	user = userRows[0]

	return
}
