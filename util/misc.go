package util

import (
	"os"
	"path/filepath"
	"strings"
)

var installPath string

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

func RemoveEmpty(toChange string) string {
	symbs := []string{" ", "\n", "\t", "\r"}
	for _, symb := range symbs {
		toChange = strings.Replace(toChange, symb, "", -1)
	}
	return toChange
}
