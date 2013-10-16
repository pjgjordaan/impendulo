package mongo

import (
	//	"fmt"
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
	Importer string
)

func ExportData(db, outFile string, cols []string) error {
	files := make(map[string][]byte, len(cols))
	for _, col := range cols {
		outFile := filepath.Join(os.TempDir(), col+".json")
		res := tool.RunCommand([]string{"mongoexport", "-d", db, "-c", col, "-o", outFile}, nil)
		if res.Err != nil {
			return res.Err
		}
		f, err := os.Open(outFile)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		files[col+".json"] = data
	}
	zipped, err := util.ZipMap(files)
	if err != nil {
		return err
	}
	return util.SaveFile(outFile, zipped)
}

func ImportData(db string, zipData []byte) error {
	tmpDir := filepath.Join(os.TempDir(), strconv.FormatInt(time.Now().Unix(), 10))
	err := util.Unzip(tmpDir, zipData)
	if err != nil {
		return err
	}
	dbImporter := Importer(db)
	filepath.Walk(tmpDir, dbImporter.ImportFile)
	return nil
}

func (this Importer) ImportFile(path string, info os.FileInfo, inErr error) (err error) {
	if inErr != nil {
		err = inErr
		return
	}
	if !strings.HasSuffix(path, ".json") {
		return
	}
	elems := strings.Split(filepath.Base(path), ".")
	if len(elems) != 2 {
		err = fmt.Errorf("Invalid collection file %s.", path)
		return
	}
	col := elems[0]
	res := tool.RunCommand([]string{"mongoimport", "-d", string(this), "-c", col, "--file", path}, nil)
	if res.Err != nil {
		return res.Err
	}
	return nil
}
