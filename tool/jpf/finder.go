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
	base, err := util.BaseDir()
	if err != nil {
		return
	}
	path := filepath.Join(base, fname)
	classes, err = loadClasses(path)
	if err == nil {
		return
	}
	data, err := findClasses(tipe, path)
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
	//Load configurations
	finderDir, err := config.JPF_FINDER.Path()
	if err != nil {
		return
	}
	home, err := config.JPF_HOME.Path()
	if err != nil {
		return
	}
	gson, err := config.GSON.Path()
	if err != nil {
		return
	}
	java, err := config.JAVA.Path()
	if err != nil {
		return
	}
	//Setup and compile JPFFinder
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
	//Run JPFFinder and retrieve the output.
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
