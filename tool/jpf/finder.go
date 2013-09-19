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

package jpf

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var (
	listenersFile = "listeners.json"
	searchesFile  = "searches.json"
)

type (
	//Class represents properties of a Java class, specifically its name and package.
	Class struct {
		Name    string
		Package string
	}
)

//Listeners retrieves all JPF Listener classes.
func Listeners() ([]*Class, error) {
	return GetClasses("listeners", listenersFile)
}

//Searches retrieves all JPF Search classes.
func Searches() ([]*Class, error) {
	return GetClasses("searches", searchesFile)
}

//GetClasses retrieves an array of classes matching a specific type and writes them to a
//provided output file for future use.
func GetClasses(tipe, fname string) (classes []*Class, err error) {
	var data []byte
	path := filepath.Join(util.BaseDir(), fname)
	classes, err = loadClasses(path)
	if err == nil {
		return
	}
	data, err = findClasses(tipe, path)
	if err != nil {
		return
	}
	classes, err = readClasses(data)
	return
}

//findClasses searches for classes in the jpf-core directory tree which match
//a specific type using JPFFinder, a Java class which searches for all concrete subclasses
//of a class or interface (gov.nasa.jpf.search.Search or gov.nasa.jpf.JPFListener for example).
//These classes are then written to a Json output file.
func findClasses(tipe, fname string) (found []byte, err error) {
	finderDir, err := config.Directory(config.JPF_FINDER)
	if err != nil {
		return
	}
	home, err := config.Directory(config.JPF_HOME)
	if err != nil {
		return
	}
	gson, err := config.JarFile(config.GSON)
	if err != nil {
		return
	}
	java, err := config.Binary(config.JAVA)
	if err != nil {
		return
	}
	target := tool.NewTarget("JPFFinder.java", "java", "finder", finderDir)
	cp := filepath.Join(home, "build", "main") + ":" + target.Dir + ":" + gson
	comp, err := javac.New(cp)
	if err != nil {
		return
	}
	_, err = comp.Run(bson.NewObjectId(), target)
	if err != nil {
		return
	}
	args := []string{java, "-cp", cp, target.Executable(), tipe, fname}
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(fname)
	if err == nil {
		found = util.ReadBytes(resFile)
	} else if execRes.Err != nil {
		err = execRes.Err
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run finder: %q.",
			string(execRes.StdErr))
	}
	return
}

//readClasses unmarshalls an array of type *Class from a Json byte array.
func readClasses(data []byte) (classes []*Class, err error) {
	err = json.Unmarshal(data, &classes)
	return
}

//loadClasses loads an array of type *Class from a Json file.
func loadClasses(fname string) (vals []*Class, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	data := util.ReadBytes(f)
	vals, err = readClasses(data)
	return
}
