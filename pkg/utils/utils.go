package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopx.io/gopx-api/pkg/log"
)

// LogError logs non-nill error at level Error.
func LogError(err error) {
	if err != nil {
		log.Error("%s", err)
	}
}

// LogWarn logs non-nill error at level Warn.
func LogWarn(err error) {
	if err != nil {
		log.Warn("%s", err)
	}
}

// Exists checks whether the path exists or not.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// IsExistingDir checks whether the path is an existing direcotory or not.
func IsExistingDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return false, nil
		default:
			return false, err
		}
	}
	return fi.Mode().IsDir(), nil
}

// IsExistingFile checks whether the path is an existing file or not.
func IsExistingFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return false, nil
		default:
			return false, err
		}
	}
	return fi.Mode().IsRegular(), nil
}

// IsDir checks whether the fi represents a directory.
func IsDir(fi os.FileInfo) bool {
	return fi.Mode().IsDir()
}

// IsFile checks whether the fi represents a regular file.
func IsFile(fi os.FileInfo) bool {
	return fi.Mode().IsRegular()
}

// IsSymlink checks whether the fi represents a symlink.
func IsSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

// CompressTarGZ compresses the source directory in .tar.gz format
// and writes to the w writer.
func CompressTarGZ(src string, w io.Writer, comment string) error {
	src, err := filepath.Abs(src)
	if err != nil {
		return errors.Wrapf(err, "Couldn't find absolute path of source dir")
	}

	dirExists, err := IsExistingDir(src)
	if err != nil {
		return errors.Wrapf(err, "Unable to tar source dir")
	}

	if !dirExists {
		return fmt.Errorf("Source dir doesn't exist: %s", src)
	}

	gzw, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
	if err != nil {
		return errors.Wrapf(err, "Unable to tar source dir")
	}
	gzw.Comment = comment
	gzw.ModTime = time.Now()
	gzw.Name = filepath.Base(src)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	err = filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Tar only following ones:
		// 1. Regular files,
		// 2. Directories,
		// 3. Symbolic link files
		if !(IsFile(fi) || IsDir(fi) || IsSymlink(fi)) {
			return nil
		}

		var link string
		if IsSymlink(fi) {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, src)
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		// Use forward slash (/) for relative path in tar instead of os path Separator
		relPath = strings.Replace(relPath, string(filepath.Separator), "/", -1)

		if header.Typeflag == tar.TypeDir {
			header.Name = fmt.Sprintf("%s/", relPath)
		} else {
			header.Name = relPath
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return errors.Wrapf(err, "Unable to write tar header for: %s", relPath)
		}

		if header.Typeflag != tar.TypeReg {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(err, "Unable to open file: %s", path)
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		if err != nil {
			return errors.Wrapf(err, "Unable to write file entry: %s", relPath)
		}

		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "Unable to tar source dir")
	}

	return nil
}

// DecompressTarGZ decompresses a .tar.gz file into
// the destination directory and returns the comment header.
func DecompressTarGZ(dst string, r io.Reader) (string, error) {
	dst, err := filepath.Abs(dst)
	if err != nil {
		return "", errors.Wrapf(err, "Couldn't find absolute path for dest dir")
	}

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to create the destination dir")
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer gzr.Close()
	comment := gzr.Comment

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err != nil {
			switch err {
			case io.EOF:
				return comment, nil
			default:
				return "", err
			}
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", errors.Wrapf(err, "Unable to create file for entry: %s", header.Name)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", errors.Wrapf(err, "Unable to read entry: %s", header.Name)
			}
			f.Close()
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", errors.Wrapf(err, "Unable to create dir for entry: %s", header.Name)
			}
		case tar.TypeSymlink:
			if targetExists, err := Exists(target); err != nil {
				return "", errors.Wrapf(err, "Unable to create symlink for entry: %s", header.Name)
			} else if targetExists {
				// If target symlink already exists, do nothing
			} else {
				err := os.Symlink(header.Linkname, target)
				if err != nil {
					return "", errors.Wrapf(err, "Unable to create symlink for entry: %s", header.Name)
				}
			}
		}
	}
}
