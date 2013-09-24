//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//Package config provides mechanisms for configuring the location of various resources
//required by Impendulo such as tools and tool configurations.
package config

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/util"
	"os"
	"path/filepath"
)

type (
	//Configuration is used to store Impendulo's Json configuration.
	Configuration struct {
		Bin map[Bin]string
		Cfg map[Cfg]string
		Dir map[Dir]string
		Jar map[Jar]string
		Sh  map[Sh]string
	}
	//Bin is a string type for paths to binary files.
	Bin string
	//Cfg is a string type for paths to configuration files.
	Cfg string
	//Dir is a string type for paths to directories.
	Dir string
	//Jar is a string type for paths to jar files.
	Jar string
	//Sh is a string type for paths to scripts.
	Sh string

	//ConfigError is used to create configuration errors.
	ConfigError struct {
		msg string
	}
)

const (
	//Executables
	DIFF  Bin = "diff"
	JAVA  Bin = "java"
	JAVAC Bin = "javac"

	//Configurations
	CHECKSTYLE_CFG Cfg = "checkstyle_cfg"
	PMD_CFG        Cfg = "pmd_cfg"

	//Directories
	JPF_FINDER    Dir = "jpf_finder"
	JPF_HOME      Dir = "jpf_home"
	JPF_RUNNER    Dir = "jpf_runner"
	JUNIT_TESTING Dir = "junit_testing"

	//Jars
	ANT        Jar = "ant"
	ANT_JUNIT  Jar = "ant_junit"
	CHECKSTYLE Jar = "checkstyle"
	FINDBUGS   Jar = "findbugs"
	GSON       Jar = "gson"
	JPF        Jar = "jpf"
	JPF_RUN    Jar = "jpf_run"
	JUNIT      Jar = "junit"

	//Scripts
	DIFF2HTML Sh = "diff2html"
	PMD       Sh = "pmd"
)

var (
	config      *Configuration
	defaultFile string
)

func init() {
	err := LoadConfigs(DefaultConfig())
	if err != nil {
		util.Log(err)
	}
}

//DefaultConfig retrieves the default configuration file path.
//This is $IMPENDULO_PATH/config/config.json
func DefaultConfig() string {
	if defaultFile != "" {
		return defaultFile
	}
	defaultFile = filepath.Join(util.InstallPath(), "config", "config.json")
	return defaultFile
}

//LoadConfigs loads configurations from a file.
//The configuration file is in Json format and looks as follows:
//
//Config: {
//    configuration_type_1      :{
//                                 "name": "value",
//                                 ...
//                                 "another_name": "another_value"
//                              },
//   ...
//   another_configuration_type :{
//                                "some_name": "some_value",
//                                ...
//                                "a_name": "a_value"
//                              }
//}
//
//Supported configuration types are currently:
//binaries (bin), configs (cfg), directories (dir), jars (jar) and scripts (sh).
func LoadConfigs(fname string) (err error) {
	//Load configuration from Json file.
	cfgFile, err := os.Open(fname)
	if err != nil {
		return
	}
	data := util.ReadBytes(cfgFile)
	err = json.Unmarshal(data, &config)
	if err != nil {
		return
	}
	//Check if configurations are valid.
	for _, bin := range config.Bin {
		if !util.IsExec(bin) {
			return Invalid("executable file", bin)
		}
	}
	for _, cfg := range config.Cfg {
		if !util.IsFile(cfg) {
			return Invalid("file", cfg)
		}
	}
	for _, dir := range config.Dir {
		if !util.IsDir(dir) {
			return Invalid("directory", dir)
		}
	}
	for _, jar := range config.Jar {
		if !util.IsFile(jar) {
			return Invalid("file", jar)
		}
	}
	for _, sh := range config.Sh {
		if !util.IsExec(sh) {
			return Invalid("executable file", sh)
		}
	}
	return
}

//Binary attempts to retrieve the named binary's path.
func Binary(name Bin) (ret string, err error) {
	ret, ok := config.Bin[name]
	if !ok {
		err = NA(name)
	}
	return
}

//Config attempts to retrieve the named config's path.
func Config(name Cfg) (ret string, err error) {
	ret, ok := config.Cfg[name]
	if !ok {
		err = NA(name)
	}
	return
}

//Directory attempts to retrieve the named directory's path.
func Directory(name Dir) (ret string, err error) {
	ret, ok := config.Dir[name]
	if !ok {
		err = NA(name)
	}
	return
}

//JarFile attempts to retrieve the named jar's path..
func JarFile(name Jar) (ret string, err error) {
	ret, ok := config.Jar[name]
	if !ok {
		err = NA(name)
	}
	return
}

//Script attempts to retrieve the named script's path..
func Script(name Sh) (ret string, err error) {
	ret, ok := config.Sh[name]
	if !ok {
		err = NA(name)
	}
	return
}

//NA creates a new ConfigError if a request is made for an unavailable configuration.
func NA(name interface{}) error {
	return &ConfigError{
		msg: fmt.Sprintf("Config %s not found.", name),
	}
}

//Invalid creates a new ConfigError for a bad configuration specification.
func Invalid(tipe, name string) error {
	return &ConfigError{
		msg: fmt.Sprintf("Bad configuration: %s is not a %s.", tipe, name),
	}
}

//Error
func (err *ConfigError) Error() string {
	return err.msg
}
