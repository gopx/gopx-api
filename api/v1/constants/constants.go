package constants

import (
	"github.com/pkg/errors"
)

// Constants holds maximum size of pagination on a single query.
const (
	UsersQueryMaxPageSize    = 100
	PackagesQueryMaxPageSize = 100
)

// Constants holds field names on which search term can be applied.
var (
	UserQueryIns         = []string{"username", "name", "email"}
	UserQueryInsDbMap    = []string{"username", "name", "email"}
	PackageQueryIns      = []string{"name", "desc", "tag"}
	PackageQueryInsDbMap = []string{"name", "description", "tag"}
)

// Constants for user specific query order.
var (
	UserSortByCols       = []string{"joined", "packages", "username", "id"}
	UserSortByDbColsMap  = []string{"joined_at", "packages_count", "username", "id"}
	UserDefaultSortByCol = UserSortByCols[0]
)

// Constants for package specific query order.
var (
	PackageSortByCols       = []string{"created", "downloads", "updated", "name", "id"}
	PackageSortByDbColsMap  = []string{"published_at", "downloads", "last_released_at", "name", "id"}
	PackageDefaultSortByCol = PackageSortByCols[0]
)

// Constants for sorting.
var (
	SortOrders       = []string{"ASC", "DESC"}
	DefaultSortOrder = SortOrders[0]
)

// Error constants
var (
	ErrInternalServer = errors.New("Internal server error occurred")
)

// PackageNameMaxLength is the maximum allowed length of a package.
const PackageNameMaxLength = 300

// TempFileNamePrefixForPackageRegister is the prefix for temp file name
// while registering new package.
const TempFileNamePrefixForPackageRegister = "gopx-pkg-upload-data-"

// PackageDataMaxSize is the maximum size of a package allowed in GoPx registry.
const PackageDataMaxSize = int64(200 * 1024 * 1024)

// PackageMetaFileNames holds the possible file names consisting of GoPx package metadata.
var PackageMetaFileNames = []string{"gopx.json", "gopx.yaml", "gopx.yml"}

// PackageUploadParamName is the param name should be given in package uploading request.
const PackageUploadParamName = "data"
