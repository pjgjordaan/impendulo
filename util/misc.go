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
	"github.com/godfried/impendulo/util/errors"

	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
)

//empty
type (
	E   struct{}
	Set map[string]E
)

var (
	baseDir, installPath string
)

func NewSet(vals ...string) Set {
	s := make(map[string]E)
	for _, v := range vals {
		s[v] = E{}
	}
	return s
}

func (s Set) String() string {
	str := ""
	for k, _ := range s {
		str += " " + k
	}
	return "{" + str + " }"
}

func (s Set) Contains(k string) bool {
	_, ok := s[k]
	return ok
}

func (s Set) Add(k string) {
	if !s.Contains(k) {
		s[k] = E{}
	}
}

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
	p := os.Getenv("GOPATH")
	if p == "" {
		return "", errors.GoPath
	}
	installPath = filepath.Join(p, "src", "github.com", "godfried", "impendulo")
	return installPath, nil
}

//BaseDir retrieves the Impendulo directory.
func BaseDir() (string, error) {
	if baseDir != "" {
		return baseDir, nil
	}
	c, e := user.Current()
	if e != nil {
		return "", e
	}
	d := filepath.Join(c.HomeDir, ".impendulo")
	if Exists(d) {
		baseDir = d
		return baseDir, nil
	}
	if e = os.MkdirAll(d, DPERM); e != nil {
		return "", e
	}
	baseDir = d
	return baseDir, nil
}

//RemoveEmpty removes whitespace characters from a string.
func RemoveEmpty(c string) string {
	return RemoveAll(c, " ", "\n", "\t", "\r")
}

func RemoveAll(c string, symbols ...string) string {
	for _, s := range symbols {
		c = strings.Replace(c, s, "", -1)
	}
	return c
}

//Or returns true if i is equal to any of the members of as.
func Or(i interface{}, as ...interface{}) bool {
	for _, a := range as {
		if i == a {
			return true
		}
	}
	return false
}

//And returns true if i is equal to all of the members of as.
func And(i interface{}, as ...interface{}) bool {
	for _, a := range as {
		if i != a {
			return false
		}
	}
	return true
}

//ShortName gets the shortened class name of a Java class.
func ShortName(n string) string {
	return Shorten(n, ".", 2)
}

func ClassName(n string) string {
	return Shorten(n, ".", 1)
}

func Shorten(n, sep string, p int) string {
	e := strings.Split(n, sep)
	c := len(e)
	if c < 2 {
		return n
	}
	return strings.Join(e[c-p:], sep)
}

func Title(s string) string {
	if len(s) < 2 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
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

func PackageClass(exec string) (string, string) {
	s := strings.Split(exec, ".")
	switch len(s) {
	case 0, 1:
		return "", exec
	default:
		return strings.Join(s[0:len(s)-1], "."), s[len(s)-1]
	}
}
