package jpf

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
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

func FindListeners() ([]byte, error) {
	target := tool.NewTarget("ListenerFinder.java", "java", "listener", config.GetConfig(config.LISTENER_DIR))
	cp := filepath.Join(config.GetConfig(config.JPF_HOME), "build", "main") + ":" + target.Dir + ":" + config.GetConfig(config.GSON_JAR)
	compArgs := []string{config.GetConfig(config.JAVAC), "-cp", cp, target.FilePath()}
	execArgs := []string{config.GetConfig(config.JAVA), "-cp", cp, target.Executable()}
	_, stderr, err := tool.RunCommand(compArgs...)
	if err != nil {
		return nil, err
	} else if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not compile listener finder: %q.", string(stderr))
	}
	stdout, stderr, err := tool.RunCommand(execArgs...)
	if err != nil {
		return nil, err
	} else if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not run listener finder: %q.", string(stderr))
	}
	return stdout, err
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
