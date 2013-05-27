package tool

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
	"os"
	"github.com/godfried/cabanga/util"
	"path/filepath"
)

func TestGetArgs(t *testing.T) {
	fb := &Tool{bson.NewObjectId(), "findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", "warnings", []string{"java", "-jar"}, []string{"-textui", "-low"}, bson.M{}, PKG_PATH}
	javac := &Tool{bson.NewObjectId(), "compile", "java", "javac", "warnings", "errors", []string{}, []string{"-implicit:class"}, bson.M{"-cp": ""}, FILE_PATH}
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
	javac.setArgFlags(map[string]string{"-cp": "there"})
	res = javac.GetArgs("here")
	if !reflect.DeepEqual(compExp, res) {
		t.Error("Arguments not computed correctly", res)
	}

}

func TestSetFlagArgs(t *testing.T){
	javac := &Tool{bson.NewObjectId(), "compile", "java", "javac", "warnings", "errors", []string{}, []string{"-implicit:class"}, bson.M{"-cp": ""}, FILE_PATH}
	expected := bson.M{"-cp":"there"}
	javac.setArgFlags(map[string]string{"-cp":"there"})
	if !reflect.DeepEqual(expected, javac.ArgFlags){
		t.Error("Flags not set properly", expected, javac.ArgFlags)
	}
}

func TestRunCommand(t *testing.T){
	failCmd := []string{"chmod", "777"}
	_, _, ok, err := RunCommand(failCmd...)
	if err == nil{
		t.Error("Command should have failed", err)
	}
	succeedCmd := []string{"ls","-a","-l"}
	_, _, ok, err = RunCommand(succeedCmd...)
	if !ok || err != nil{
		t.Error(err)
	}
	noCmd := []string{"lsa"}
	_, _, ok, err = RunCommand(noCmd...)
	if ok{
		t.Error("Command should not have started", err)
	}
}


func TestRun(t *testing.T){
	fileId := bson.NewObjectId()
javac := &Tool{bson.NewObjectId(), "compile", "java", "javac", "warnings", "errors", []string{}, []string{"-implicit:class"}, bson.M{"-cp": ""}, FILE_PATH}
	ti, err := setupTarget()
	if err != nil{
		t.Error(err)
	}
	_, err = javac.Run(fileId, ti, map[string]string{"-cp":ti.Dir})
	if err != nil{
		t.Error(err)
	}
	os.RemoveAll(ti.Dir)
}

	

func setupTarget() (*TargetInfo, error){
	var fileData = []byte(`package bermuda;

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

