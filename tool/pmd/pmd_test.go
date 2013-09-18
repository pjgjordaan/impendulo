package pmd

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
)

var (
	file = []byte(`
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
)

func TestRun(t *testing.T) {
	location := filepath.Join(os.TempDir(), "Triangle")
	srcLocation := filepath.Join(location, "triangle")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	os.Mkdir(srcLocation, util.DPERM)
	target := tool.NewTarget("Triangle.java",
		tool.JAVA, "triangle", location)
	err := util.SaveFile(target.FilePath(), file)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	rules, err := DefaultRules(bson.NewObjectId())
	if err != nil {
		t.Error(err)
	}
	pmd := New(rules)
	_, err = pmd.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	os.Remove(filepath.Join(location, "pmd.xml"))
	pmd = New(nil)
	res, err := pmd.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Error(err)
	}
	pmdCfg := config.Config(config.PMD)
	defer config.SetConfig(config.PMD, pmdCfg)
	config.SetConfig(config.PMD, "")
	pmd = New(rules)
	res, err = pmd.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Errorf("Expected error, got %s.", res)
	}
}
