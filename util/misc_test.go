package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallPath(t *testing.T) {
	loc := "here"
	os.Clearenv()
	err := os.Setenv("IMPENDULO_PATH", loc)
	if err != nil {
		t.Error(err)
	}
	path, err := InstallPath()
	if err != nil {
		t.Error(err)
	}
	if path != loc {
		t.Error(fmt.Errorf("Invalid Installpath %s", path))
	}
	installPath = ""
	os.Clearenv()
	_, err = InstallPath()
	if err == nil {
		t.Error(errors.New("Expected error GOPATH not set."))
	}
	err = os.Setenv("GOPATH", loc)
	if err != nil {
		t.Error(err)
	}
	installPath = ""
	path, err = InstallPath()
	if err != nil {
		t.Error(err)
	}
	loc = filepath.Join(loc, "src", "github.com", "godfried", "impendulo")
	if path != loc {
		t.Error(fmt.Errorf("Invalid Installpath %s", path))
	}
}

func TestShortName(t *testing.T) {
	tests := map[string]string{
		"za.ac.sun.builder.BuilderFactory": "builder.BuilderFactory",
		"Nopackage":                        "Nopackage",
		"single.Package":                   "single.Package",
		".":                                ".",
		".EmptyPackage":                    ".EmptyPackage",
	}
	for full, short := range tests {
		if shortened := ShortName(full); shortened != short {
			t.Error(fmt.Errorf("Expected %s got %s.", short, shortened))
		}
	}

}
