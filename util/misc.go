//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package util

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	installPath string
)

//InstallPath retrieves the location where Impendulo is currently installed.
//It first checks for the IMPENDULO_PATH environment variable otherwise the
//install path is constructed from GOPATH and the Impendulo's package.
func InstallPath() string {
	if installPath != "" {
		return installPath
	}
	installPath = os.Getenv("IMPENDULO_PATH")
	if installPath != "" {
		return installPath
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH is not set.")
	}
	installPath = filepath.Join(gopath, "src",
		"github.com", "godfried", "impendulo")
	return installPath
}

//RemoveEmpty removes whitespace characters from a string.
func RemoveEmpty(toChange string) string {
	symbs := []string{" ", "\n", "\t", "\r"}
	for _, symb := range symbs {
		toChange = strings.Replace(toChange, symb, "", -1)
	}
	return toChange
}

//EqualsOne returns true if test is equal to any of the members of args.
func EqualsOne(test interface{}, args ...interface{}) bool {
	for _, arg := range args {
		if test == arg {
			return true
		}
	}
	return false
}
