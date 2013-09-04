package javac

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
	location := filepath.Join(os.TempDir(), "triangle")
	target := tool.NewTarget("Triangle.java",
		tool.JAVA, "", location)
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	err := util.SaveFile(target.FilePath(), file)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	comp := New("")
	_, err = comp.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	javac := config.Config(config.JAVAC)
	defer config.SetConfig(config.JAVAC, javac)
	config.SetConfig(config.JAVAC, "")
	comp2 := New("")
	_, err = comp2.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Errorf("Expected error.")
	}
	err = util.SaveFile(target.FilePath(), file2)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	_, err = comp.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Errorf("Expected error.")
	}
	target = tool.NewTarget("File.java",
		tool.JAVA, "", location)
	_, err = comp.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Error("Expected error")
	}

}

var file = []byte(`
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

var file2 = []byte(`
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

`)
