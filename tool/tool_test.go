package tool

import (
	"github.com/godfried/impendulo/util"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCommand(t *testing.T) {
	failCmd := []string{"chmod", "777"}
	_, _, err := RunCommand(failCmd, nil)
	if err == nil {
		t.Error("Command should have failed", err)
	}
	succeedCmd := []string{"ls", "-a", "-l"}
	_, _, err = RunCommand(succeedCmd, nil)
	if err != nil {
		t.Error(err)
	}
	noCmd := []string{"lsa"}
	_, _, err = RunCommand(noCmd, nil)
	if _, ok := err.(*StartError); !ok {
		t.Error("Command should not have started", err)
	}
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
	ti := NewTarget(fname, "java", pkg, dir)
	err := util.SaveFile(ti.FilePath(), fileData)
	if err != nil {
		return nil, err
	}
	return ti, nil
}
