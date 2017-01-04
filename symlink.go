// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"os"
	"golang.org/x/sys/windows"
	"encoding/binary"
)
 
// From the Cygwin user guide (https://www.cygwin.com/cygwin-ug-net/using.html):
//
//    The default symlinks are plain files containing a
//    magic cookie followed by the path to which the link
//    points. They are marked with the DOS SYSTEM attribute
//    so that only files with that attribute have to be
//    read to determine whether or not the file is a
//    symbolic link.
//
//    Note:
//    Cygwin symbolic links are using UTF-16 to encode the filename
//    of the target file, to better support internationalization.
//    Symlinks created by old Cygwin releases can be read just fine.
//    However, you could run into problems with them if you're now
//    using another character set than the one you used when creating
//    these symlinks (see the section called “Potential Problems when
//    using Locales”).
var symlinkMagicCookie = []byte(`!<symlink>`)

func cyglink(fn string, target string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(symlinkMagicCookie)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte{0xff, 0xfe}) // UTF16 BOM (little endian)
	if err != nil {
		return err
	}

	targetutf16, err := windows.UTF16FromString(target)
	if err != nil {
		return err
	}

	err = binary.Write(f, binary.LittleEndian, targetutf16)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte{0, 0})
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	winfn, err := windows.UTF16PtrFromString(fn)
	if err != nil {
		return err
	}

	err = windows.SetFileAttributes(winfn, windows.FILE_ATTRIBUTE_SYSTEM)
	if err != nil {
		return err
	}

	return nil
}
