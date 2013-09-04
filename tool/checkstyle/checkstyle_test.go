package checkstyle

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
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
	check := New()
	_, err = check.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	os.Remove(filepath.Join(location, "checkstyle.xml"))
	checkCfg := config.Config(config.CHECKSTYLE)
	defer config.SetConfig(config.CHECKSTYLE, checkCfg)
	config.SetConfig(config.CHECKSTYLE, "")
	check2 := New()
	res, err := check2.Run(bson.NewObjectId(), target)
	if err == nil {
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
