package jpf

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var listenersFile = "listeners.gob"

func Listeners() (listeners []*Listener, err error) {
	var data []byte
	path := filepath.Join(util.BaseDir(), listenersFile)
	listeners, err = loadListeners(path)
	if err == nil {
		return
	}
	data, err = FindListeners()
	if err != nil {
		return
	}
	listeners, err = readListeners(data)
	if err != nil {
		return
	}
	err = saveListeners(listeners, path)
	return
}

func FindListeners() (found []byte, err error) {
	target := tool.NewTarget("ListenerFinder.java", "java", "listener", config.GetConfig(config.LISTENER_DIR))
	cp := filepath.Join(config.GetConfig(config.JPF_HOME), "build", "main") + ":" + target.Dir + ":" + config.GetConfig(config.GSON_JAR)
	comp := javac.NewJavac(cp)
	_, err = comp.Run(bson.NewObjectId(), target)
	if err != nil {
		return
	}
	args := []string{config.GetConfig(config.JAVA), "-cp", cp, target.Executable()}
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		err = execRes.Err
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run listener finder: %q.", string(execRes.StdErr))
	}
	found = execRes.StdOut
	return
}

type Listener struct {
	Name    string
	Package string
}

func readListeners(data []byte) (listeners []*Listener, err error) {
	err = json.Unmarshal(data, &listeners)
	return
}

func saveListeners(vals []*Listener, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating file %q", err, fname)
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&vals)
	if err != nil {
		return fmt.Errorf("Encountered error %q while encoding map %q to file %q", err, vals, fname)
	}
	return nil
}

func loadListeners(fname string) (vals []*Listener, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(f)
	err = dec.Decode(&vals)
	return
}
