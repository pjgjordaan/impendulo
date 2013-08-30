package pmd

import (
	"testing"
	"os"
	"labix.org/v2/mgo/bson"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/config"
	"path/filepath"
)

func TestRun(t *testing.T){
	location := filepath.Join(os.TempDir(), "Triangle")
	srcLocation := filepath.Join(location, "triangle")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	os.Mkdir(srcLocation, util.DPERM)
	target := tool.NewTarget("Triangle.java", 
		project.JAVA, "triangle", location)
	err := util.SaveFile(target.FilePath(), file)
	if err != nil{
		t.Errorf("Could not save file %q", err)
	}
	pmd := New(GetRules())
	_, err = pmd.Run(bson.NewObjectId(), target)
	if err != nil{
		t.Errorf("Expected success, got %q", err)
	}
	os.Remove(filepath.Join(location, "pmd.xml"))
	pmd = New(nil)
	res, err := pmd.Run(bson.NewObjectId(), target)
	if err == nil{
		t.Errorf("Expected error, got %s.", res)
	}
	pmdCfg := config.GetConfig(config.PMD)
	defer config.SetConfig(config.PMD, pmdCfg)
	config.SetConfig(config.PMD, "")
	pmd = New(GetRules())
	res, err = pmd.Run(bson.NewObjectId(), target)
	if err == nil{
		t.Errorf("Expected error, got %s.", res)
	}
}

var file = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)