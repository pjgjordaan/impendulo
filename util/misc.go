//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package util

import (
	"errors"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	baseDir, installPath string
)

//InstallPath retrieves the location where Impendulo is currently installed.
//It first checks for the IMPENDULO_PATH environment variable otherwise the
//install path is constructed from GOPATH and the Impendulo's package.
func InstallPath() (string, error) {
	if installPath != "" {
		return installPath, nil
	}
	installPath = os.Getenv("IMPENDULO_PATH")
	if installPath != "" {
		return installPath, nil
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.New("GOPATH is not set.")
	}
	installPath = filepath.Join(
		gopath, "src", "github.com",
		"godfried", "impendulo",
	)
	return installPath, nil
}

//BaseDir retrieves the Impendulo directory.
func BaseDir() (string, error) {
	if baseDir != "" {
		return baseDir, nil
	}
	cur, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cur.HomeDir, ".impendulo")
	if Exists(dir) {
		baseDir = dir
		return baseDir, nil
	}
	err = os.MkdirAll(dir, DPERM)
	if err != nil {
		return "", err
	}
	baseDir = dir
	return baseDir, nil
}

//RemoveEmpty removes whitespace characters from a string.
func RemoveEmpty(toChange string) string {
	return RemoveAll(toChange, " ", "\n", "\t", "\r")
}

func RemoveAll(toChange string, symbols ...string) string {
	for _, symb := range symbols {
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

//ShortName gets the shortened class name of a Java class.
func ShortName(exec string) string {
	elements := strings.Split(exec, ".")
	num := len(elements)
	if num < 2 {
		return exec
	}
	return strings.Join(elements[num-2:], ".")
}

func Min(a, b int) int {
	if a > b {
		a = b
	}
	return a
}

func Max(a, b int) int {
	if a < b {
		a = b
	}
	return a
}

func Title(s string) string {
	if len(s) < 2 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func Round(x float64, prec int) float64 {
	p := math.Pow(10, float64(prec))
	r := x * p
	if r < 0.0 {
		r -= 0.5
	} else {
		r += 0.5
	}
	return float64(int64(r)) / p
}
func SplitTitles(titles string) []string {
	var a []string
	i := 0
	for s := titles; s != ""; s = s[i:] {
		i = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if i <= 0 {
			i = len(s)
		}
		a = append(a, s[:i])
	}
	return a
}
