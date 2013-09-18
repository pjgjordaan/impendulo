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

package tool

import (
	"path/filepath"
	"strings"
)

type (
	//TargetInfo stores information about the target file.
	TargetInfo struct {
		Name    string
		Lang    string
		Package string
		Ext     string
		Dir     string
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
func NewTarget(name, lang, pkg, dir string) *TargetInfo {
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
		Lang:    lang,
		Package: pkg,
		Ext:     ext,
		Dir:     dir,
	}
}
