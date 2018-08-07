package handler

import (
	"strings"

	"github.com/pkg/errors"
	"gopx.io/gopx-api/api/v1/auth"
	"gopx.io/gopx-api/api/v1/constants"
	"gopx.io/gopx-api/api/v1/controller/user"
	"gopx.io/gopx-common/str"
)

func authUser(authValue string) (u *user.QueryRow, err error) {
	authType, err := auth.Parse(authValue)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse the auth value")
		return
	}

	whereClause := ""
	args := []interface{}{}

	switch v := authType.(type) {
	case *auth.AuthenticationTypeBasic:
		whereClause = "username = ? and password = ?"
		args = append(args, v.Username(), auth.CreateHash(v.Password()))
	case *auth.AuthenticationTypeAPIKey:
		whereClause = "api_key = ?"
		args = append(args, auth.CreateHash(v.APIKey()))
	default:
		err = errors.Errorf("Auth type %s is not supported yet", v.Name())
		return
	}

	sortBy := "id ASC"
	limit := "1"

	userRows, err := user.Query(whereClause, sortBy, limit, "", args...)
	if err != nil {
		err = constants.ErrInternalServer
		return
	}

	if len(userRows) >= 1 {
		u = userRows[0]
	}

	return
}

func listOsNames(os string) []string {
	os = strings.TrimSpace(os)
	if str.IsEmpty(os) {
		return []string{}
	}

	oParts, _ := str.Split(os, "\\s*,\\s*")
	osList := []string{}
	for _, v := range oParts {
		if str.IsEmpty(v) {
			continue
		}

		aParts, _ := str.Split(v, "\\s*\\:\\s*")
		osArch := aParts[0]
		if len(aParts) >= 2 {
			if !str.IsEmpty(aParts[1]) {
				osArch += ":" + aParts[1]
			}
		}

		osList = append(osList, osArch)
	}
	return osList
}
