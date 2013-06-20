package tool

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGenericGetArgs(t *testing.T) {
	fb := &GenericTool{"findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", []string{"java", "-jar"}, []string{"-textui", "-low"}, map[string]string{}, PKG_PATH}
	javac := &GenericTool{"javac", "java", "javac", []string{}, []string{"-implicit:class"}, map[string]string{"-cp": ""}, FILE_PATH}
	fbExp := []string{"java", "-jar", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "-textui", "-low", "here"}
	res := fb.GetArgs("here")
	if !reflect.DeepEqual(fbExp, res) {
		t.Error("Arguments not computed correctly", res)
	}
	compExp := []string{"javac", "-implicit:class", "-cp", "there", "here"}
	res = javac.GetArgs("here")
	if reflect.DeepEqual(compExp, res) {
		t.Error("Arguments not computed correctly", res)
	}
	javac.AddArgs(map[string]string{"-cp": "there"})
	res = javac.GetArgs("here")
	if !reflect.DeepEqual(compExp, res) {
		t.Error("Arguments not computed correctly", res)
	}
}

func TestAddArgs(t *testing.T) {
	javac := &GenericTool{"javac", "java", "javac", []string{}, []string{"-implicit:class"}, nil, FILE_PATH}
	expected := map[string]string{"-cp": "there"}
	javac.AddArgs(expected)
	if !reflect.DeepEqual(expected, javac.args) {
		t.Error("Flags not set properly", expected, javac.args)
	}
}

func TestRunCommand(t *testing.T) {
	failCmd := []string{"chmod", "777"}
	_, _, ok, err := RunCommand(failCmd...)
	if err == nil {
		t.Error("Command should have failed", err)
	}
	succeedCmd := []string{"ls", "-a", "-l"}
	_, _, ok, err = RunCommand(succeedCmd...)
	if !ok || err != nil {
		t.Error(err)
	}
	noCmd := []string{"lsa"}
	_, _, ok, err = RunCommand(noCmd...)
	if ok {
		t.Error("Command should not have started", err)
	}
}

func TestGenericRun(t *testing.T) {
	fileId := bson.NewObjectId()
	javac := &GenericTool{"javac", "java", "javac", []string{}, []string{"-implicit:class"}, nil, FILE_PATH}
	ti, err := setupTarget()
	if err != nil {
		t.Error(err)
	}
	javac.AddArgs(map[string]string{"-cp": ti.Dir})
	_, err = javac.Run(fileId, ti)
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(ti.Dir)
}

func setupTarget() (*TargetInfo, error) {
	fileData := []byte(`package bermuda;

public class Triangle {
        public int maxpath(int[][] tri) {
                int h = tri.length;
                for (int j = h - 2; j >= 0; j--) {
                        for (int i = 0; i <= j; i++) {
                                tri[i][j] = Math.max(tri[i + 1][j], tri[i + 1][j + 1]);
                        }
                }
                return tri[0][0];
        }
}`)
	fname := "Triangle.java"
	pkg := "bermuda"

	dir := filepath.Join(os.TempDir(), "test")
	ti := NewTarget("Triangle", fname, "java", pkg, dir)
	err := util.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), fileData)
	if err != nil {
		return nil, err
	}
	return ti, nil
}
