// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"github.com/ulikunitz/xz"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func checkFn(fn string) bool {
	splat := strings.Split(fn, "/")
	for _, elem := range splat {
		if elem == "." || elem == ".." {
			return false
		}
	}
	return true
}

func extractTo(absFn string, targetDir string) error {
	var tr *tar.Reader = nil

	f, err := os.Open(absFn)
	if err != nil {
		return err
	}
	defer f.Close()

	// gzip
	if strings.HasSuffix(absFn, ".tar.gz") || strings.HasSuffix(absFn, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		tr = tar.NewReader(gzr)
		// bz2
	} else if strings.HasSuffix(absFn, ".tar.bz2") || strings.HasSuffix(absFn, ".tbz") {
		tr = tar.NewReader(bzip2.NewReader(f))
		// xz
	} else if strings.HasSuffix(absFn, ".tar.xz") {
		xzr, err := xz.NewReader(f)
		if err != nil {
			return err
		}
		tr = tar.NewReader(xzr)
		// lzma, etc?
	} else {
		log.Fatal("unhandled file extension: %v", absFn)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if !checkFn(hdr.Name) {
			log.Fatal("bad fn in tarball: %v", hdr.Name)
		}

		// Map /usr/bin -> /bin
		if strings.HasPrefix(hdr.Name, "usr/bin") {
			hdr.Name = strings.Replace(hdr.Name, "usr/bin", "bin", 1)
		}
		// Map /usr/lib -> /lib
		if strings.HasPrefix(hdr.Name, "usr/lib") {
			hdr.Name = strings.Replace(hdr.Name, "usr/lib", "lib", 1)
		}

		if hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA {
			f, err := os.Create(filepath.Join(targetDir, hdr.Name))
			if err != nil {
				return err
			}

			defer f.Close()

			_, err = io.Copy(f, tr)
			if err != nil {
				return err
			}
		} else if hdr.Typeflag == tar.TypeDir {
			err = os.MkdirAll(filepath.Join(targetDir, hdr.Name), 0750)
			if err != nil {
				return err
			}
		} else if hdr.Typeflag == tar.TypeLink {
			err = os.Link(filepath.Join(targetDir, hdr.Linkname), filepath.Join(targetDir, hdr.Name))
			if err != nil {
				return err
			}
		} else if hdr.Typeflag == tar.TypeSymlink {
			err = cyglink(filepath.Join(targetDir, hdr.Name), hdr.Linkname)
			if err != nil {
				return err
			}
		} else {
			log.Fatalf("fatal error: unhandled typeflag %v", hdr.Typeflag)
		}
	}

	return nil
}
