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
