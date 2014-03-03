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

package tool

import (
	"github.com/godfried/impendulo/util"
	"path/filepath"
	"strings"
)

type (
	//Target stores information about the target file.
	Target struct {
		Name    string
		Package string
		Ext     string
		Dir     string
		Lang    Language
	}
)

//FilePath
func (t *Target) FilePath() string {
	return filepath.Join(t.PackagePath(), t.FullName())
}

//PackagePath
func (t *Target) PackagePath() string {
	if t.Package != "" {
		vals := strings.Split(t.Package, ".")
		return filepath.Join(t.Dir, filepath.Join(vals...))
	} else {
		return t.Dir
	}
}

//FullName
func (t *Target) FullName() string {
	return t.Name + "." + t.Ext
}

//Executable retrieves the path to the compiled executable with its package.
func (t *Target) Executable() string {
	if t.Package != "" {
		return t.Package + "." + t.Name
	} else {
		return t.Name
	}
}

//NewTarget
func NewTarget(name, pkg, dir string, lang Language) *Target {
	name, ext := util.Extension(name)
	return &Target{
		Name:    name,
		Package: pkg,
		Ext:     ext,
		Dir:     dir,
		Lang:    lang,
	}
}
