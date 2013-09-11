package jpf

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
	location := filepath.Join(os.TempDir(), "Racer")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	target := tool.NewTarget("Racer.java",
		tool.JAVA, "", location)
	err := util.SaveFile(target.FilePath(), srcFile)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	jpfConfig := NewConfig(bson.NewObjectId(), "user", jpfBytes)
	jpf, err := New(jpfConfig, location)
	if err != nil {
		t.Errorf("Could not load jpf %q", err)
	}
	_, err = jpf.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	jpfCfg := config.Config(config.JPF_JAR)
	defer config.SetConfig(config.JPF_JAR, jpfCfg)
	config.SetConfig(config.JPF_JAR, "")
	jpf, err = New(jpfConfig, location)
	if err == nil {
		t.Error("Expected error.")
	}
}

var srcFile = []byte(`
public class Racer implements Runnable {

     int d = 42;

     public void run () {
          doSomething(1001);                   // (1)
          d = 0;                               // (2)
     }

     public static void main (String[] args){
          Racer racer = new Racer();
          Thread t = new Thread(racer);
          t.start();

          doSomething(1000);                   // (3)
          int c = 420 / racer.d;               // (4)
          System.out.println(c);
     }
     
     static void doSomething (int n) {
          // not very interesting..
          try { Thread.sleep(n); } catch (InterruptedException ix) {}
     }
}
`)

var jpfBytes = []byte(`
listener=gov.nasa.jpf.listener.PreciseRaceDetector
`)
