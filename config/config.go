//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

//Package config provides mechanisms for configuring how impendulo should be run.
package config

import (
	"bufio"
	"fmt"
	"github.com/godfried/impendulo/util"
	"os"
	"path/filepath"
	"strings"
)

var (
	settings      map[string]string
	defaultConfig string
)

const (
	JUNIT_EXEC        = "junit_exec"
	LINT4J            = "lint4j"
	FINDBUGS          = "findbugs"
	JUNIT_JAR         = "junit_jar"
	ANT               = "ant"
	ANT_JUNIT         = "ant_junit"
	JAVAC             = "javac"
	JAVA              = "java"
	JPF_JAR           = "jpf_jar"
	RUNJPF_JAR        = "runjpf_jar"
	GSON_JAR          = "gson_jar"
	JPF_RUNNER_DIR    = "jpf_runner_dir"
	TESTING_DIR       = "testing_dir"
	JPF_HOME          = "jpf_home"
	JPF_FINDER_DIR    = "jpf_finder_dir"
	PMD               = "pmd"
	PMD_RULES         = "pmd_rules"
	CHECKSTYLE        = "checkstyle"
	CHECKSTYLE_CONFIG = "checkstyle_config"
	DIFF2HTML         = "diff2html"
	DIFF              = "diff"
)

func init() {
	settings = make(map[string]string)
	err := LoadConfigs(DefaultConfig())
	if err != nil {
		util.Log(err)
	}
}

//DefaultConfig
func DefaultConfig() string {
	if defaultConfig != "" {
		return defaultConfig
	}
	defaultConfig = filepath.Join(util.InstallPath(), "config.txt")
	return defaultConfig
}

//LoadConfigs loads configurations from a file.
//Configurations are key-value pairs on different lines.
//Keys are seperated from the value by a '='.
func LoadConfigs(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
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

//Config attempts to retrieve the named config.
func Config(name string) string {
	ret, ok := settings[name]
	if !ok {
		panic("Config not found: " + name)
	}
	return ret
}

//SetConfig sets the config 'name' to 'value'.
func SetConfig(name, value string) {
	settings[name] = value
}
