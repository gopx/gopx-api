package helper

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"gopx.io/gopx-api/api/v1/constants"
)

// SanitizePackageVersion parses the version string and returns the sanitized
// form for database.
func SanitizePackageVersion(v string) (sVersion string, err error) {
	oVer, err := semver.NewVersion(v)
	if err != nil {
		err = errors.New("The package version should be in semvar format")
		return
	}

	sVersion = fmt.Sprintf(
		"%d.%d.%d",
		oVer.Major(),
		oVer.Minor(),
		oVer.Patch(),
	)

	if oVer.Prerelease() != "" {
		sVersion = fmt.Sprintf("%s-%s", sVersion, oVer.Prerelease())
	}

	return
}

// ValidatePackageName checks whether the package name
// meets the naming constraints.
func ValidatePackageName(name string) error {
	ln := utf8.RuneCountInString(name)
	if !(ln > 0 && ln < constants.PackageNameMaxLength) {
		return errors.Errorf("Package name must be non-empty and maximum %d characters long", constants.PackageNameMaxLength)
	}

	if strings.ToLower(name) != name {
		return errors.Errorf("Package name must contain only lowercase charactars")
	}

	if url.PathEscape(name) != name {
		return errors.Errorf("Package name must not contain any non-url-safe character")
	}

	if matched, err := regexp.MatchString(`[@.~\\/!'()*\s]`, name); err != nil {
		return constants.ErrInternalServer
	} else if matched {
		return errors.Errorf(`Package name must not contain any of these special characters: @, ., ~, \, /, !, ', (, ), *`)
	}

	return nil
}

// IsSameVersion checks wether the two semvar versions are same or not.
func IsSameVersion(v1, v2 string) (ok bool, err error) {
	oV1, err := semver.NewVersion(v1)
	if err != nil {
		return
	}

	oV2, err := semver.NewVersion(v2)
	if err != nil {
		return
	}

	ok = oV1.Equal(oV2)

	return
}
