package auth

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
	"gopx.io/gopx-common/str"
)

const (
	authTypeBasic  = "Basic"
	authTypeAPIKey = "APIKey"
)

// AuthenticationType represents the http request auth type.
type AuthenticationType interface {
	Name() string
}

// AuthenticationTypeBasic represents the Basic http auth type.
type AuthenticationTypeBasic struct {
	name     string
	username string
	password string
}

// Username returns the username value.
func (atb *AuthenticationTypeBasic) Username() string {
	return atb.username
}

// Password returns the auth password.
func (atb *AuthenticationTypeBasic) Password() string {
	return atb.password
}

// Name returns auth type name.
func (atb *AuthenticationTypeBasic) Name() string {
	return atb.name
}

// AuthenticationTypeAPIKey represents the API Key http auth type.
type AuthenticationTypeAPIKey struct {
	name   string
	apiKey string
}

// APIKey returns the API Key value.
func (ata *AuthenticationTypeAPIKey) APIKey() string {
	return ata.apiKey
}

// Name returns auth type name.
func (ata *AuthenticationTypeAPIKey) Name() string {
	return ata.name
}

// AuthenticationTypeUnknown represents an unrecognized http auth type.
type AuthenticationTypeUnknown struct {
	name string
}

// Name returns auth type name.
func (atu *AuthenticationTypeUnknown) Name() string {
	return atu.name
}

func parseBasicAuth(auth string) (username, password string, err error) {
	bytes, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return
	}

	auth = string(bytes)
	idx := strings.IndexRune(auth, ':')
	if idx == -1 {
		username = auth
		password = ""
	} else {
		username = auth[:idx]
		password = auth[idx+1:]
	}

	return
}

// Parse parses the http Authorization header value and returns
// the corresponding auth type.
func Parse(auth string) (authType AuthenticationType, err error) {
	auth = strings.TrimSpace(auth)
	parts := str.SplitSpace(auth)

	if len(parts) < 2 {
		err = errors.New("Invalid auth data")
		return
	}

	switch aType, aVal := parts[0], parts[1]; aType {
	case authTypeBasic:
		u, p, err := parseBasicAuth(aVal)
		if err != nil {
			err = errors.Wrap(err, "Invalid basic auth base64 value")
			return nil, err
		}
		authType = &AuthenticationTypeBasic{
			name:     aType,
			username: u,
			password: p,
		}
	case authTypeAPIKey:
		authType = &AuthenticationTypeAPIKey{
			name:   aType,
			apiKey: aVal,
		}
	default:
		authType = &AuthenticationTypeUnknown{
			name: aType,
		}
	}

	return
}

// CreateHash creates the hash form of the input value.
func CreateHash(value string) string {
	return value
}
