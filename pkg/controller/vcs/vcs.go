package vcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"gopx.io/gopx-api/pkg/config"
	errorCtrl "gopx.io/gopx-api/pkg/controller/error"
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
	// PackageTypePrivate indicates the package is private.
	PackageTypePrivate
)

// PackageMeta represents the package meta info required for vcs registry.
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

// RegisterPackage registers a new package to the vcs registry.
func RegisterPackage(meta *PackageMeta, data io.Reader) (err error) {
	reqPath := "/v1/packages"
	reqMethod := http.MethodPost

	metaBuff := bytes.Buffer{}
	err = json.NewEncoder(&metaBuff).Encode(meta)
	if err != nil {
		err = errors.Wrap(err, "Failed to encode meta data to JSON")
		return
	}

	rBody, wBody := io.Pipe()
	mW := multipart.NewWriter(wBody)

	req, err := http.NewRequest(reqMethod, createReqURL(reqPath), rBody)
	if err != nil {
		err = errors.Wrap(err, "Failed to create the request")
		return
	}
	req.Header.Set("Authorization", createAuthHeader())
	req.Header.Set("Content-Type", mW.FormDataContentType())
	req.Header.Set("User-Agent", "GoPx API Service")

	go func() {
		var err error
		defer func() {
			wBody.CloseWithError(err)
		}()

		metaPart, err := mW.CreateFormField("meta")
		if err != nil {
			err = errors.Wrap(err, "Failed to create part for meta field")
			return
		}

		_, err = io.Copy(metaPart, &metaBuff)
		if err != nil {
			err = errors.Wrap(err, "Failed to write meta data to part")
			return
		}

		pkgDataPart, err := mW.CreateFormFile("data", fmt.Sprintf("%s.tar.gz", meta.Name))
		if err != nil {
			err = errors.Wrap(err, "Failed to create part for package data field")
			return
		}

		_, err = io.Copy(pkgDataPart, data)
		if err != nil {
			err = errors.Wrap(err, "Failed to write package data to part")
			return
		}

		err = mW.Close()
		if err != nil {
			err = errors.Wrap(err, "Failed to finish writing of multipart data")
			return
		}
	}()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrap(err, "Failed to send the request")
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			err = errors.Wrap(err, "Failed to read the response")
			return err
		}
		msg, err := errorCtrl.DecodeErrorMessage(body)
		if err != nil {
			err = errors.Wrap(err, "Failed to decode error response")
			return err
		}

		return errors.Errorf("Response code %d: %s", res.StatusCode, msg)
	}

	return
}

// DeletePackage removes package data from vcs registry.
func DeletePackage(packageName string) (err error) {
	reqPath := fmt.Sprintf("/v1/packages/%s", packageName)
	reqMethod := http.MethodDelete

	req, err := http.NewRequest(reqMethod, createReqURL(reqPath), http.NoBody)
	if err != nil {
		err = errors.Wrap(err, "Failed to create the request")
		return
	}
	req.Header.Set("Authorization", createAuthHeader())
	req.Header.Set("User-Agent", "GoPx API Service")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrap(err, "Failed to send the request")
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			err = errors.Wrap(err, "Failed to read the response")
			return err
		}
		msg, err := errorCtrl.DecodeErrorMessage(body)
		if err != nil {
			err = errors.Wrap(err, "Failed to decode error response")
			return err
		}

		return errors.Errorf("Response code %d: %s", res.StatusCode, msg)
	}

	return

}

func createReqURL(path string) string {
	var (
		vcsServiceHost = os.Getenv(config.Env.GoPxVCSAPIIP)
		vcsServicePort = os.Getenv(config.Env.GoPxVCSAPIPort)
	)

	oURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(vcsServiceHost, vcsServicePort),
		Path:   path,
	}

	return oURL.String()
}

func createAuthHeader() string {
	return fmt.Sprintf("AuthKey %s", os.Getenv(config.Env.GoPxVCSAPIAuthKey))
}
