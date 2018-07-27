package types

// Package represents a single GoPX package.
type Package struct {
	Name           string `json:"name"`
	Desc           string `json:"desc"`
	Downloads      uint64 `json:"downloads"`
	TotalDownloads uint64 `json:"totalDownloads"`
}
