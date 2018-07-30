package helper

// PaginationConfig holds the pagination related configurations.
type PaginationConfig struct {
	Page         uint64
	PerPageCount uint64
}

// SortingConfig holds the sorting related configurations.
type SortingConfig struct {
	SortBy string
	Order  string
}
