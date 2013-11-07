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
	"path/filepath"
	"strings"
)

type (
	//TargetInfo stores information about the target file.
	TargetInfo struct {
		Name    string
		Package string
		Ext     string
		Dir     string
		Lang    Language
	}
)

//FilePath
func (ti *TargetInfo) FilePath() string {
	return filepath.Join(ti.PackagePath(), ti.FullName())
}

//PackagePath
func (ti *TargetInfo) PackagePath() string {
	if ti.Package != "" {
		vals := strings.Split(ti.Package, ".")
		return filepath.Join(ti.Dir, filepath.Join(vals...))
	} else {
		return ti.Dir
	}
}

//FullName
func (ti *TargetInfo) FullName() string {
	return ti.Name + "." + ti.Ext
}

//Executable retrieves the path to the compiled executable with its package.
func (ti *TargetInfo) Executable() string {
	if ti.Package != "" {
		return ti.Package + "." + ti.Name
	} else {
		return ti.Name
	}
}

//NewTarget
func NewTarget(name, pkg, dir string, lang Language) *TargetInfo {
	split := strings.Split(name, ".")
	var ext string
	if len(split) < 2 {
		ext = ""
	} else {
		name = split[0]
		ext = split[1]
	}
	return &TargetInfo{
		Name:    name,
		Package: pkg,
		Ext:     ext,
		Dir:     dir,
		Lang:    lang,
	}
}
