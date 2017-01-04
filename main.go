// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"os"
	"log"
	"io"
	"fmt"
	"errors"
	"path/filepath"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
)

var (
	cygwinArch string = "x86"
	cygwinMirrors []string = []string{
		"http://mirrors.dotsrc.org/cygwin",
	}
	cygwinPackages []string = []string{
		"base-cygwin",
		"cygwin",
		"base-files", 
		"bash",
		"patch",
		"tar",
		"xz",
		"gzip",
		"bzip2",
		"hostname",
		"curl",
		"which",
		"unzip",
	}
)

func distfilePath(args ...string) string {
	path := []string{Args.Distfiles()}
	path = append(path, args...)
	return filepath.Join(path...)
}

func checkSHA512(absFn string, expectedLength int64, insha512 string) error {
	hw := sha512.New512_256()

	f, err := os.Open(absFn)
	if err != nil {
		return err
	}

	defer f.Close()

	n, err := io.Copy(hw, f)	
	if err != nil {
		return err
	}

	if n != expectedLength {
		return fmt.Errorf("Length mismsatch for '%v'. Has %v, want %v", absFn, n, expectedLength)
	}

	sumBytes := hw.Sum(nil)
	sumHex := hex.EncodeToString(sumBytes)

	if sumHex != insha512 {
		return fmt.Errorf("SHA512 mismatch for %v. Has %v, want %v", absFn, sumHex, insha512)
	}

	return nil
}

func ensureDownloaded(mirrorRelativeURL string, fileSize int64, sha512sum string) error {
	mirrors := []string{}
	for _, mirror := range cygwinMirrors {
		mirrors = append(mirrors, mirror)
	}

	for {
		if len(mirrors) == 0 {
			return errors.New("no remaining mirrors")
		}

		// Take the front-most mirror.
		mirrorBase := mirrors[0]
		mirrors = mirrors[1:]

		url := mirrorBase + "/" + mirrorRelativeURL
		outFn := distfilePath(mirrorRelativeURL)
		dir := filepath.Dir(outFn)
		err := os.MkdirAll(dir, 0755)
		if os.IsExist(err) {
			// All directories we require already exist. All good.
		} else if err != nil {
			log.Fatalf("unable to mkdir: %v", err)
		}
		
		rsp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer rsp.Body.Close()
		
		newf, err := os.Create(outFn)
		if err != nil {
			return err
		}
		defer newf.Close()

		_, err = io.Copy(newf, rsp.Body)
		if err != nil {
			return err
		}

		if len(sha512sum) == 0 {
			return nil
	 	} else {
			if fileSize == -1 {
				return errors.New("If ensureDownloaded is passed a sha512sum, it must also be passed a fileSize.")
			}
			return checkSHA512(outFn, fileSize, sha512sum)
		}
	}

	panic("unreachable")
}

func prepareTarget(targetDir string) error {
	_, err := os.Stat(Args.Target)
	if err == nil {
		return fmt.Errorf("Target directory '%v' already exists. Aborting.", Args.Target)
	}

	dirs := []string{
		targetDir,
		filepath.Join(targetDir, "usr"),
		filepath.Join(targetDir, "tmp"),
		filepath.Join(targetDir, "dev"),
	}

	for _, absDir := range dirs {
		err := os.MkdirAll(absDir, 0750)
		if os.IsExist(err) {
			// Great, it already exists...
		} else if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := ParseArgs()
	if err != nil {
		log.Fatalf("unable to parse args: %v", err)
	}

	if !Args.FetchOnly {
		log.Printf("preparing target '%v'", Args.Target)
		err = prepareTarget(Args.Target)
		if err != nil {
			log.Fatalf("prepareTarget failed: %v", err)
		}
	}
	
	log.Printf("preparing distfiles '%v'", Args.Distfiles())
	err = os.MkdirAll(Args.Distfiles(), 0755)
	if os.IsExist(err) {
		// OK...
	} else if err != nil {
		log.Fatalf("unable to prepare distfiles: %v", err)
	}

	log.Printf("fetching setup.ini")
	err = ensureDownloaded(Args.Arch + "/setup.ini", -1, "")
	if err != nil {
		log.Fatalf("unable to download setup.ini: %v", err)
	}

	log.Printf("fetching setup.ini.sig")
	err = ensureDownloaded(Args.Arch + "/setup.ini.sig", -1, "")
	if err != nil {
		log.Fatalf("unable to downlaod setup.ini.sig: %v", err)
	}

	log.Printf("verifying setup.ini.sig")
	err = verifySetupIniSignature(Args.Arch + "/setup.ini")
	if err != nil {
		log.Fatalf("unable to verify setup.ini signature: %v", err)
	}

	log.Printf("reading setup.ini")
	dist, err := parseSetupIni(Args.Arch + "/setup.ini")
	if err != nil {
		log.Fatalf("unable to parse setup.ini: %v", err)
	}

	// Install all requested packages.
	for _, pkg := range cygwinPackages {
		log.Printf("installing package '%v'", pkg)
		err = installPkg(pkg, dist, Args.Target, nil)
		if err == ErrAlreadyInstalled {
			// Might already be installed via a requirement... That's fine.
		} else if err != nil {
			log.Fatalf("package install failed: %v", err)
		}
	}

	if !Args.KeepDistfiles {
		log.Printf("removing distfiles directory")
		err = os.RemoveAll(Args.Distfiles())
		if err != nil {
			log.Fatalf("unable to remove distfiles directory: %v", err)
		}
	}

	log.Printf("done")
}
