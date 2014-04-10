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

//Package mongo is used to export/import collections from/to a mongo database.
package mongo

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	//Importer is used to walk a directory containing mongodb collections stored in json
	//files and then import these collections into the database specified by this Importer.
	Importer string
)

//ExportData exports data from collections in the specified database
//to a specified location as a zip file. This makes use of the mongoexport utility.
func ExportData(outFile, db string, cols []string) error {
	fs := make(map[string][]byte, len(cols))
	for _, col := range cols {
		o := filepath.Join(os.TempDir(), col+".json")
		_, e := tool.RunCommand([]string{"mongoexport", "-d", db, "-c", col, "-o", o}, nil)
		if e != nil {
			return e
		}
		f, e := os.Open(o)
		if e != nil {
			return e
		}
		d, e := ioutil.ReadAll(f)
		if e != nil {
			return e
		}
		fs[col+".json"] = d
	}
	zs, e := util.ZipMap(fs)
	if e != nil {
		return e
	}
	return util.SaveFile(outFile, zs)
}

//ImportData imports collections stored in a zip file
//to the specified database.
func ImportData(db string, zip []byte) error {
	td := filepath.Join(os.TempDir(), strconv.FormatInt(time.Now().Unix(), 10))
	if e := util.Unzip(td, zip); e != nil {
		return e
	}
	filepath.Walk(td, Importer(db).ImportFile)
	return nil
}

//ImportFile imports a single collection found in the file specified by path
//to the database specified by this Importer. This makes use of the mongoimport utility.
func (i Importer) ImportFile(path string, info os.FileInfo, inErr error) error {
	if inErr != nil {
		return inErr
	}
	if !strings.HasSuffix(path, ".json") {
		return nil
	}
	sp := strings.Split(filepath.Base(path), ".")
	if len(sp) != 2 {
		return fmt.Errorf("invalid collection file %s", path)
	}
	_, e := tool.RunCommand([]string{"mongoimport", "-d", string(i), "-c", sp[0], "--file", path}, nil)
	return e
}
