// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"strings"
	"os"
	"os/exec"
	"path/filepath"
	"errors"
)

var (
	ErrAlreadyInstalled = errors.New("package already installed")
)

func installPkg(name string, dist *Distribution, targetDir string, excludeRequirements []string) error {
	if dist.IsInstalled(name) {
		return ErrAlreadyInstalled
	}

	pkg, err := dist.Get(name)
	if err != nil {
		return err
	} 

	requirements := pkg.Requirements()
	for _, req := range requirements {
		if strings.HasPrefix(req, "_") {
			// Skip internal hint
			continue
		}
		excluded := false
		for _, excludedReq := range excludeRequirements {
			if req == excludedReq {
				excluded = true
				break
			}
		}
		if excluded {
			// Skip excluded requirement.
			continue
		}
		if !dist.IsInstalled(req) {
			newExcludeRequirements := []string{}
			newExcludeRequirements = append(newExcludeRequirements, excludeRequirements...)
			newExcludeRequirements = append(newExcludeRequirements, name)
			installPkg(req, dist, targetDir, newExcludeRequirements)
		}
	}

	relativeUrl, fileSize, sha512sum := pkg.InstallInfo()

	ensureDownloaded(relativeUrl, fileSize, sha512sum)
	absFn := distfilePath(relativeUrl)

	if !Args.FetchOnly {
		err = os.MkdirAll(targetDir, 0750) 
		if os.IsExist(err) {
			// OK...
		} else if err != nil {
			return err
		}

		err = extractTo(absFn, targetDir)
		if err != nil {
			return err
		}

		postInstallDir := filepath.Join(targetDir, "etc", "postinstall")

		f, err := os.Open(postInstallDir)
		if err != nil {
			return err
		}

		defer f.Close()

		entries, err := f.Readdirnames(-1)
		if err != nil {
			return err
		}

		for _, fn := range entries {
			absFn := filepath.Join(postInstallDir, fn)
			if strings.HasSuffix(absFn, ".done") {
				continue
			}
			bashExe := filepath.Join(targetDir, "bin", "bash.exe")
			if _, err := os.Stat(bashExe); err == nil {
				env := os.Environ()
				newenv := []string{}
				for _, envvar := range env {
					if strings.HasPrefix(strings.ToLower(envvar), "path=") {
						newenv = append(newenv, "PATH=" + filepath.Join(targetDir, "bin") + ";" + os.Getenv("PATH"))
					} else {
						newenv = append(newenv, envvar)
					}
				}
				cmd := exec.Command(bashExe, "--norc", "--noprofile", absFn)
				cmd.Env = newenv
				_, err := cmd.CombinedOutput()
				if err != nil {
					return err
				}

				err = os.Rename(absFn, absFn + ".done")
				if err != nil {
					return err
				}
			}
		}
	}

	dist.MarkAsInstalled(name)

	return nil
}
