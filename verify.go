// Copyright 2005-2017 The Mumble Developers. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file at the root of the
// Mumble source tree or at <https://www.mumble.info/LICENSE>.

package main

import (
	"os"
	"strings"
	"golang.org/x/crypto/openpgp"
)

// Cygwin public keyring via https://cygwin.com/key/pubring.asc
var cygwinPubring = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.2.6 (GNU/Linux)

mQGiBEhS+m8RBAC5bn3n2yG0eqNlpg/D7DkZXQfFUBZN1D4sL/NsXKISQkA3FsiT
enDYDMFCy3NJiCDcihJprP2xs4Fc25MEcmJ4j9X93bCV4DtHv22qO1XWGkxr/XQJ
ZxYmUxFhezBOCZd+wXir0izIsGghR1+ei6i+vL4mRYy8wpMCKwf8X0qRywCg1l2J
a91PsTO6itVUACYMvKNFCHED/RenUG+kYRch9YHuDwG9LxkhgwSEZ0NIGUgZLHMY
HZDlcWBRoV6uPcqa2iKs8vvAENMcGWqo+fuRycGQ6+zlFn29IoHrcxMMM27VpifQ
91N5AqgSMPOIFkKse2VNFQ2jL4t1NfdQazRvZojwkXuYY9kB16h0Y2Zme1Pt5RgC
/wLhA/4lkttrs3ElzkAOZtrTwi7tCJnNR8/5VYnVd63NEGyAXk/qralUoQO+GNQf
ZXJUvCoYIhinHh7vzfqMT2l1gGi0FuSULX3dY5jsm0Vcu+f7XLlDoEurx1vDYCv+
9QABQDDPXuZJk55pDG1TQbvAFV8U6wWdCI5hBwcJsDfwLMzxN7QaQ3lnd2luIDxj
eWd3aW5AY3lnd2luLmNvbT6IXgQTEQIAHgUCSFL6bwIbAwYLCQgHAwIDFQIDAxYC
AQIeAQIXgAAKCRCpomL/Z2BBuncZAKCmfQS2ROcl9H8VaKmdMOB/loNRLwCfTqxf
W6L6ifl1uDwoH8t83PRjkRW5AQ0ESFL6cBAEAIqcw0vcqdTvuukm6oiRUxkQ/jrP
+4w2FNKEK1sYG5+cbwVrf3ISTUrbTRbV3Fz5npefwaLNlIUjVYCBBWL4PuUtL4cC
rmbvMXabSYfz2qg/aqqw9xNa4G9GCdF4j9AIZaV86UHElC1wZAHTvMEdgHs8ek9k
b5rDDChUgyE+nXQ7AAMFA/4rXq6swR8m/1O8nRgNkwDvas3DbUOIdoYoFPrN7e2L
BuYWFDB+O2IUn6tAgHhDxpzO9vw58U5a/z1zm63Lf9ybHDV4c3Rqie2u2oberj1K
KStnn27KlGGvFY9kWe9WKh+ZN90/oqVGBT4+obmTiwUmVJIUy4vSZDjC0VqZHLxd
OIhJBBgRAgAJBQJIUvpwAhsMAAoJEKmiYv9nYEG6euAAniloWCmYSp4ULCHauEMb
opO2jFlwAKCwlu0FsfcO/2+AresM67hCSwxQ+g==
=7qzL
-----END PGP PUBLIC KEY BLOCK-----
`

func verifySetupIniSignature(setupIniRelativeURL string) error {
	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(cygwinPubring))
	if err != nil {
		return err
	}

	setupIniPath := distfilePath(setupIniRelativeURL)
	f, err := os.Open(setupIniPath)
	if err != nil {
		return err
	}

	defer f.Close()

	setupIniSigPath := distfilePath(setupIniRelativeURL + ".sig")
	sig, err := os.Open(setupIniSigPath)
	if err != nil {
		return err
	}

	defer sig.Close()

	_, err = openpgp.CheckDetachedSignature(keyring, f, sig)
	if err != nil {
		return err
	}

	return nil
}
