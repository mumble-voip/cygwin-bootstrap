// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"errors"
	"os"
	"bufio"
	"strings"
	"io"
	"fmt"
	"strconv"
)

type Distribution struct {
	Arch string
	Packages []Package

	InstalledPackages map[string]bool
}

func NewDistribution() *Distribution {
	d := new(Distribution)
	d.Packages = make([]Package, 0)
	d.InstalledPackages = make(map[string]bool)
	return d
}

func (dist *Distribution) Get(name string) (Package, error) {
	for _, pkg := range dist.Packages {
		if pkg.Name() == name {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("no such package: %v", name)
}

func (dist *Distribution) IsInstalled(name string) bool {
	if _, ok := dist.InstalledPackages[name]; ok {
		return true
	}
	return false
}

func (dist *Distribution) MarkAsInstalled(name string) {
	dist.InstalledPackages[name] = true
}

type Package struct {
	Meta map[string]interface{}
}

func (pkg *Package) Name() string {
	if name, ok := pkg.Meta["name"]; ok {
		return name.(string)
	}
	return ""
}

func (pkg *Package) Requirements() []string {
	if reqs, ok := pkg.Meta["requires"]; ok {
		switch v := reqs.(type) {
		case []string:
			return v
		case string:
			return []string{v}
		}
	}
	return []string{}
}
 
func (pkg *Package) InstallInfo() (relativeUrl string, fileSize int64, sha512sum string) {
	if val, ok := pkg.Meta["install"]; ok {
		info := val.([]string)
		fsz, err := strconv.ParseInt(info[1], 10, 64)
		if err != nil {
			fsz = int64(-1)
		}

		relativeUrl = info[0]
		fileSize = fsz
		sha512sum = info[2]

		return
	}
	return "", -1, ""
}

func parseSetupIni(setupIniRelativeURL string) (*Distribution, error) {
	setupIniPath := distfilePath(setupIniRelativeURL)

	f, err := os.Open(setupIniPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	dist := NewDistribution()
	meta := make(map[string]interface{})
	prefix := ""

	b := bufio.NewReader(f)

	for {
		lineBuf, isPrefix, err := b.ReadLine()
		if err == io.EOF {
			break
		}
		if isPrefix {
			return nil, errors.New("isprefix unhandled")
		}
		if err != nil {
			return nil, err
		}

		line := string(lineBuf) + string("\n")

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		// Context switch
		} else if strings.HasPrefix(line, "@") {
			pkgName := strings.TrimSpace(line[1:])
			meta["name"] = pkgName
		// Sections (we seem them as prefixes)
		} else if strings.HasPrefix(line, "[") {
			trimmedLine := strings.TrimSpace(line)
			prefix = trimmedLine[1:len(trimmedLine)-1]
		// End context
		} else if strings.TrimSpace(line) == "" {
			if _, ok := meta["name"]; !ok {
				meta["name"] = "__root__"
			}
			dist.Packages = append(dist.Packages, Package{
				Meta: meta,
			})
			meta = make(map[string]interface{})
			prefix = ""
		} else {
			colon := strings.Index(line, ":")
			if colon < 0 {
				return nil, errors.New("expected colon")
			}
			key := line[0:colon]
			value := line[colon+2:]
			if line[colon+1] != ' ' {
				return nil, errors.New("expected space after colon")
			}
			if strings.Count(line, `"`) > 0 { // read quoted string
				beforeStrLit := ""
				doubleQuoteIdx := strings.Index(value, `"`)
				if doubleQuoteIdx > 1 {
					beforeStrLit = value[0:doubleQuoteIdx]
				}
				strLit := ""
				idx := doubleQuoteIdx+1
				escapeSequence := false
				getch := func() string {
					retVal := ""
					if idx >= len(value) {
						lineBuf, isPrefix, err = b.ReadLine()
						if err != nil {
							panic(err)
						}
						if isPrefix {
							panic("isPrefix")
						}
						value += string(lineBuf) + string("\n")
					}
					retVal = string(value[idx])
					idx += 1
					return retVal
				}
				for {
					c := getch()
					if c == `"` {
						if escapeSequence {
							strLit += `"`
							escapeSequence = false
							continue
						} else {
							break
						}
					} else if c == `\` {
						escapeSequence = true
						continue
					} else {
						strLit += c
					}
				}
				if len(beforeStrLit) > 0 {
					meta[prefix+key] = []string{beforeStrLit, strLit}
				} else {
					meta[prefix+key] = strLit
				}
			} else if strings.Count(value, " ") > 0 { // read list
				meta[prefix+key] = strings.Split(strings.TrimSpace(value), " ")
			} else { // read string
				meta[prefix+key] = strings.TrimSpace(value)
			}
		}
	}

	return dist, nil
}
