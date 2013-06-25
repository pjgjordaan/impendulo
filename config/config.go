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
	return settings[name]
}

func SetConfig(name, value string) {
	settings[name] = value
}

const (
	JUNIT_EXEC = "junit_exec"
	LINT4J     = "lint4j"
	FINDBUGS   = "findbugs"
	JUNIT_JAR  = "junit_jar"
	JAVAC      = "javac"
	JAVA       = "java"
)
