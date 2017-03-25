// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"os"
	"path/filepath"
)

func configureEtcNsswitch(targetDir string) error {
	nsswitch := filepath.Join(targetDir, "etc", "nsswitch.conf")
	f, err := os.OpenFile(nsswitch, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte("\ndb_home: /%H\n"))
	if err != nil {
		f.Close()
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func postSetup(targetDir string) error {
	err := configureEtcNsswitch(targetDir)
	if err != nil {
		return err
	}

	return nil
}
