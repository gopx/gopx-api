package constants

// UsersQueryPageSize holds maximum size of pagination in user query.
var UsersQueryPageSize = 100

// UserQueryIns holds user fields on which the search term will be applied.
var UserQueryIns = []string{"username", "name", "email"}

// Constants for user specific query order.
var (
	UserSortByCols       = []string{"joined", "packages", "username", "id"}
	UserSortByDbColsMap  = []string{"joined_at", "packages_count", "username", "id"}
	UserDefaultSortByCol = UserSortByCols[0]
)

// Constants for sorting.
var (
	SortOrders       = []string{"ASC", "DESC"}
	DefaultSortOrder = SortOrders[0]
)
