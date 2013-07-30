package tool

import (
	"path/filepath"
	"strings"
)

//TargetInfo stores information about the target file.
type TargetInfo struct {
	Name    string
	Lang    string
	Package string
	Ext     string
	Dir     string
}

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
	return &TargetInfo{split[0], lang, pkg, split[1], dir}
}
