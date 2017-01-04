cygwin-bootstrap
================

cygwin-bootstrap is a small Go program that can bootstrap a Cygwin installation
without using Cygwin's own `setup.exe`/`setup-x86_64.exe`.

It's primarily meant to be used by mumble-releng's
(https://github.com/mumble-voip/mumble-releng) Windows build environment for easily
installing a reproducible/point-in-time snapshot of cygwin.
