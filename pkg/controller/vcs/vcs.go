package vcs

import (
	"encoding/base64"
	"io"
)

// PackageType represents type of package in vcs registry.
type PackageType int

func (p PackageType) String() string {
	switch p {
	case 0:
		return "public"
	case 1:
		return "private"
	default:
		return "unknown"
	}
}

const (
	// PackageTypePublic indicates the package is public.
	PackageTypePublic PackageType = iota
	// PackageTypePrivate indicates the package is private,
	// requires Authentication to access it.
	PackageTypePrivate
)

// PackageMeta represents the package data required for vcs registry.
type PackageMeta struct {
	Type    PackageType  `json:"type"`
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Owner   PackageOwner `json:"owner"`
}

// PackageOwner represents the owner data required for vcs registry.
type PackageOwner struct {
	Name        string `json:"name"`
	PublicEmail string `json:"publicEmail"`
	Username    string `json:"username"`
}

// PackageReadmeData holds the package README content in base64 format.
type PackageReadmeData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    uint64 `json:"size"`
	Content string `json:"content"`
}

// RegisterPackage registers a new package to the vcs registry.
// @TODO: Add implementation.
func RegisterPackage(meta *PackageMeta, data io.Reader) (err error) {
	return
}

// PackageReadme returns the content of README.
// @TODO: Add implementation.
func PackageReadme(pkgName, version string) (readme *PackageReadmeData, err error) {

	fakeContent := []byte("Hey, I am README.")

	readme = &PackageReadmeData{
		Name:    "README.md",
		Version: version,
		Size:    uint64(len(fakeContent)),
		Content: base64.StdEncoding.EncodeToString(fakeContent),
	}

	return
}

// DeletePackage removes package data from vcs registry.
// @TODO: Add implementation.
func DeletePackage(packageName string) (err error) {
	return
}
