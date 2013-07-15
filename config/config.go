package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var settings map[string]string

func LoadConfigs(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	settings = make(map[string]string)
	for scanner.Scan() {
		vals := strings.Split(scanner.Text(), "=")
		if len(vals) != 2 {
			return fmt.Errorf("Config file not formatted correctly.")
		}
		name := strings.TrimSpace(vals[0])
		value := strings.TrimSpace(vals[1])
		settings[name] = value
	}
	return scanner.Err()
}

func GetConfig(name string) string {
	ret, ok := settings[name]
	if !ok {
		panic("Config not found: " + name)
	}
	return ret
}

func SetConfig(name, value string) {
	settings[name] = value
}

const (
	JUNIT_EXEC   = "junit_exec"
	LINT4J       = "lint4j"
	FINDBUGS     = "findbugs"
	JUNIT_JAR    = "junit_jar"
	JAVAC        = "javac"
	JAVA         = "java"
	JPF_JAR      = "jpf_jar"
	RUNJPF_JAR   = "runjpf_jar"
	GSON_JAR     = "gson_jar"
	RUNNER_DIR   = "runner_dir"
	JPF_HOME     = "jpf_home"
	LISTENER_DIR = "listener_dir"
	PMD = "pmd"
	CHECKSTYLE = "checkstyle"
	CHECKSTYLE_CONFIG = "checkstyle_config"
	DIFF2HTML = "diff2html"
	DIFF = "diff"
)
