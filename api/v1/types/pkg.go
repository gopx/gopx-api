package types

import (
	"time"
)

// Package represents a single GoPx package.
type Package struct {
	Name             string    `json:"name"`
	ID               uint64    `json:"id"`
	Desc             string    `json:"desc"`
	Owner            string    `json:"owner"`
	Version          string    `json:"version"`
	Downloads        uint64    `json:"downloads"`
	PublishedAt      time.Time `json:"publishedAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	License          string    `json:"license"`
	Homepage         string    `json:"homepage"`
	RepositoryURL    string    `json:"repositoryURL"`
	DocumentationURL string    `json:"documentationURL"`
	BugsURL          string    `json:"bugsURL"`
	Engines          Engines   `json:"engines"`
	Os               []string  `json:"os"`
}

// Engines holds environment dependencies.
type Engines struct {
	Go string `json:"go"`
}

// PackageDownloads holds the downloads count for a single package.
type PackageDownloads struct {
	Name      string `json:"name"`
	ID        uint64 `json:"id"`
	Downloads uint64 `json:"downloads"`
}

// RegistryDownloads holds the overall package downloads across registry.
type RegistryDownloads struct {
	Downloads uint64 `json:"downloads"`
}

// PackageVersionHistory holds versions history of a package.
type PackageVersionHistory struct {
	Name     string           `json:"name"`
	ID       uint64           `json:"id"`
	Versions []PackageVersion `json:"versions"`
}

// PackageVersion holds info of a single version.
type PackageVersion struct {
	Version    string    `json:"version"`
	ReleasedAt time.Time `json:"releasedAt"`
}

// PackageMetaData holds the metadata of a gopx package i.e. contents of the gopx.json file.
type PackageMetaData struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Description      string                 `json:"description"`
	HomepageURL      string                 `json:"homepage"`
	Tags             []string               `json:"tags"`
	License          string                 `json:"license"`
	BugsURL          string                 `json:"bugsURL"`
	RepositoryURL    string                 `json:"repository"`
	DocumentationURL string                 `json:"docs"`
	Engines          PackageMetaDataEngines `json:"engines"`
	Os               []string               `json:"os"`
}

// PackageMetaDataEngines holds the engines metadata of a gopx package.
type PackageMetaDataEngines struct {
	Go string `json:"go"`
}

// PackageReadme holds the contents of README.
type PackageReadme struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    uint64 `json:"size"`
	Content string `json:"content"`
}
