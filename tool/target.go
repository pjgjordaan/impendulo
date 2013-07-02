package tool

import (
	"path/filepath"
	"reflect"
	"strings"
)

//TargetInfo stores information about the target file.
type TargetInfo struct {
	//	Project string
	//File name without extension
	Name string
	//Language file is written in
	Lang    string
	Package string
	Ext     string
	Dir     string
}

//FilePath
func (ti *TargetInfo) FilePath() string {
	return filepath.Join(ti.PkgPath(), ti.FullName())
}

//PkgPath
func (ti *TargetInfo) PkgPath() string {
	if ti.Package != ""{
		vals := strings.Split(ti.Package, ".")
		return filepath.Join(ti.Dir, filepath.Join(vals...))
	} else{
		return ti.Dir
	}
}

//FullName
func (ti *TargetInfo) FullName() string {
	return ti.Name + "." + ti.Ext
}

//Executable retrieves the path to the compiled executable with its package.
func (ti *TargetInfo) Executable() string {
	if ti.Package != ""{
		return ti.Package + "." + ti.Name
	} else{
		return ti.Name
	}
}

func (this *TargetInfo) Equals(that *TargetInfo) bool {
	return reflect.DeepEqual(this, that)
}

type TargetSpec int

const (
	DIR_PATH TargetSpec = iota
	PKG_PATH
	FILE_PATH
	EXEC_PATH
)

//GetTarget retrieves the target path based on the type required.
func (ti *TargetInfo) GetTarget(spec TargetSpec) string {
	switch spec {
	case DIR_PATH:
		return ti.Dir
	case PKG_PATH:
		return ti.PkgPath()
	case FILE_PATH:
		return ti.FilePath()
	case EXEC_PATH:
		return ti.Executable()
	}
	return ""
}

//NewTarget
func NewTarget(name, lang, pkg, dir string) *TargetInfo {
	split := strings.Split(name, ".")
	return &TargetInfo{split[0], lang, pkg, split[1], dir}
}
