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
	"time"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/project"
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
func GetClasses(tipe, fname string) ([]*Class, error) {
	d, e := util.BaseDir()
	if e != nil {
		return nil, e
	}
	p := filepath.Join(d, fname)
	if c, e := loadClasses(p); e == nil {
		return c, nil
	}
	data, e := findClasses(tipe, p)
	if e != nil {
		return nil, e
	}
	return readClasses(data)
}

//findClasses searches for classes in the jpf-core directory tree which match
//a specific type using JPFFinder, a Java class which searches for all concrete subclasses
//of a class or interface (gov.nasa.jpf.search.Search or gov.nasa.jpf.JPFListener for example).
//These classes are then written to a Json output file.
func findClasses(tipe, fname string) ([]byte, error) {
	//Load configurations
	fd, e := config.JPF_FINDER.Path()
	if e != nil {
		return nil, e
	}
	hd, e := config.JPF_HOME.Path()
	if e != nil {
		return nil, e
	}
	gp, e := config.GSON.Path()
	if e != nil {
		return nil, e
	}
	jp, e := config.JAVA.Path()
	if e != nil {
		return nil, e
	}
	//Setup and compile JPFFinder
	t := tool.NewTarget("JPFFinder.java", "finder", fd, project.JAVA)
	cp := filepath.Join(hd, "build", "main") + ":" + t.Dir + ":" + gp
	c, e := javac.New(cp)
	if e != nil {
		return nil, e
	}
	if _, e = c.Run(bson.NewObjectId(), t); e != nil {
		return nil, e
	}
	r, re := tool.RunCommand([]string{jp, "-cp", cp, t.Executable(), tipe, fname}, nil, 30*time.Second)
	rf, e := os.Open(fname)
	if e == nil {
		return util.ReadBytes(rf), nil
	} else if re != nil {
		return nil, re
	} else if r.HasStdErr() {
		return nil, fmt.Errorf("could not run finder: %q", string(r.StdErr))
	}
	return nil, nil
}

//readClasses unmarshalls an array of type *Class from a Json byte array.
func readClasses(data []byte) (classes []*Class, err error) {
	err = json.Unmarshal(data, &classes)
	return
}

//loadClasses loads an array of type *Class from a Json file.
func loadClasses(fname string) ([]*Class, error) {
	f, e := os.Open(fname)
	if e != nil {
		return nil, e
	}
	return readClasses(util.ReadBytes(f))
}
