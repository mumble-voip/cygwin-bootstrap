// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type args struct {
	Target              string
	DistfilesUnexpanded string
	Arch                string
	MirrorsSeparated    string
	PackagesSeparated   string
	FetchOnly           bool
	KeepDistfiles       bool
	Help                bool
}

func (a *args) Distfiles() string {
	return os.Expand(a.DistfilesUnexpanded, func(key string) string {
		switch key {
		case "target":
			return a.Target
		}
		return ""
	})
}

func (a *args) Mirrors() []string {
	return strings.Split(a.MirrorsSeparated, ",")
}

func (a *args) Packages() []string {
	return strings.Split(a.PackagesSeparated, ",")
}

var Args args

func defaultDistfiles() string {
	return filepath.Join("${target}", "distfiles")
}

func init() {
	flag.StringVar(&Args.Target, "target", "", "target directory for cygwin installation (i.e, c:\\cygwin)")
	flag.StringVar(&Args.DistfilesUnexpanded, "distfiles", defaultDistfiles(), "path where "+progName+" will store downloaded artifacts")
	flag.StringVar(&Args.Arch, "arch", "x86", "cygwin architecture (x86 or x86_64)")
	flag.StringVar(&Args.MirrorsSeparated, "mirrors", "http://mirrors.dotsrc.org/cygwin", "mirror(s) to download from (comma separated)")
	flag.StringVar(&Args.PackagesSeparated, "packages", "base-cygwin,cygwin,base-files,bash,patch,tar,xz,gzip,bzip2,hostname,curl,which,unzip,grep,gawk,vim,mingw64-i686-gcc-core,mingw64-i686-binutils,diffutils,diffstat,autoconf", "packages to install (comma separated)")
	flag.BoolVar(&Args.FetchOnly, "fetch-only", false, "only fetch distfiles, don't install anything (implies -keep-distfiles=true)")
	flag.BoolVar(&Args.KeepDistfiles, "keep-distfiles", false, "keep distfiles?")
	flag.BoolVar(&Args.Help, "help", false, "show this listing")
	flag.Parse()
}

func ParseArgs() error {
	flag.Parse()

	if Args.Help {
		fmt.Fprintf(os.Stderr, "%v v%v\n\n", progName, progVersion)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !Args.FetchOnly && Args.Target == "" {
		return fmt.Errorf("missing argument: target")
	}

	if Args.Arch != "x86" && Args.Arch != "x86_64" {
		return fmt.Errorf("invalid argument: arch is '%v' -- unknown arch!", Args.Arch)
	}

	if Args.FetchOnly {
		Args.KeepDistfiles = true
	}

	return nil
}
