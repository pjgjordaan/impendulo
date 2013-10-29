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
	//Validator is a function used to validate a file.
	Validator func(string) bool
	//File is an interface which represents a file configuration.
	File interface {
		//Valid is used to validate the provided path to the file.
		Valid(string) error
		//Description provides a description about the type of file this is.
		Description() string
		//Path retrieves the path to this file.
		Path() (string, error)
	}

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
	location, err := DefaultConfig()
	if err == nil {
		err = LoadConfigs(location)
	}
	if err != nil {
		util.Log(err)
	}
}

//DefaultConfig retrieves the default configuration file path.
//This is $IMPENDULO_PATH/config/config.json
func DefaultConfig() (string, error) {
	if defaultFile != "" {
		return defaultFile, nil
	}
	base, err := util.BaseDir()
	if err != nil {
		return "", err
	}
	defaultFile = filepath.Join(base, "config.json")
	if util.IsFile(defaultFile) {
		return defaultFile, nil
	} else if util.IsDir(defaultFile) {
		return "", fmt.Errorf("%s is a directory.", defaultFile)
	}
	iPath, err := util.InstallPath()
	if err != nil {
		fmt.Println(defaultFile, err)
		return "", err
	}
	installedConfig := filepath.Join(iPath, "config", "config.json")
	return defaultFile, util.CopyFile(defaultFile, installedConfig)
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
	iPath, err := util.InstallPath()
	if err != nil {
		return
	}
	javaPath := filepath.Join(iPath, "java")
	config.Dir[JPF_RUNNER] = filepath.Join(javaPath, "runner")
	config.Dir[JPF_FINDER] = filepath.Join(javaPath, "finder")
	config.Dir[JUNIT_TESTING] = filepath.Join(javaPath, "testing")
	//Check if configurations are valid.
	for bin, path := range config.Bin {
		if err = bin.Valid(path); err != nil {
			return
		}
	}
	for cfg, path := range config.Cfg {
		if err = cfg.Valid(path); err != nil {
			return
		}
	}
	for dir, path := range config.Dir {
		if err = dir.Valid(path); err != nil {
			return
		}
		if dir == JPF_HOME {
			jar := filepath.Join(path, filepath.Join("build", "jpf.jar"))
			if err = JPF.Valid(jar); err != nil {
				return
			}
			config.Jar[JPF] = jar
			jar = filepath.Join(path, filepath.Join("build", "RunJPF.jar"))
			if err = JPF_RUN.Valid(jar); err != nil {
				return
			}
			config.Jar[JPF_RUN] = jar
		}
	}
	for jar, path := range config.Jar {
		if err = jar.Valid(path); err != nil {
			return
		}
	}
	for sh, path := range config.Sh {
		if err = sh.Valid(path); err != nil {
			return
		}
	}
	return
}

func (cfg Cfg) Valid(path string) error {
	return valid(cfg, path, util.IsFile)
}

func (cfg Cfg) Description() string {
	return "configuration file"
}

func (cfg Cfg) Path() (string, error) {
	return path(cfg)
}

func (dir Dir) Valid(path string) error {
	return valid(dir, path, util.IsDir)
}

func (dir Dir) Description() string {
	return "directory"
}

func (dir Dir) Path() (string, error) {
	return path(dir)
}

func (jar Jar) Valid(path string) error {
	return valid(jar, path, util.IsFile)
}

func (jar Jar) Description() string {
	return "jar file"
}

func (jar Jar) Path() (string, error) {
	return path(jar)
}

func (sh Sh) Valid(path string) error {
	return valid(sh, path, util.IsExec)
}

func (sh Sh) Description() string {
	return "executable script"
}

func (sh Sh) Path() (string, error) {
	return path(sh)
}

func (bin Bin) Valid(path string) error {
	return valid(bin, path, util.IsExec)
}

func (bin Bin) Description() string {
	return "executable file"
}

func (bin Bin) Path() (string, error) {
	return path(bin)
}

//valid determines whether the provided path is valid for corresponding file.
func valid(file File, path string, validator Validator) (err error) {
	if !validator(path) {
		err = Invalid(file, path)
	}
	return
}

//path attempts to retrieve the file's path.
func path(file File) (ret string, err error) {
	var ok bool
	switch t := file.(type) {
	case Bin:
		ret, ok = config.Bin[t]
	case Cfg:
		ret, ok = config.Cfg[t]
	case Dir:
		ret, ok = config.Dir[t]
	case Jar:
		ret, ok = config.Jar[t]
	case Sh:
		ret, ok = config.Sh[t]
	default:
		ok = true
		err = &ConfigError{
			msg: fmt.Sprintf("Unknown configuration type %s.", t),
		}
	}
	if !ok {
		err = NA(file)
	}
	return
}

//NA creates a new ConfigError if a request is made for an unavailable configuration.
func NA(file File) error {
	return &ConfigError{
		msg: fmt.Sprintf("Config %s not found.", file),
	}
}

//Invalid creates a new ConfigError for a bad configuration specification.
func Invalid(file File, path string) error {
	return &ConfigError{
		msg: fmt.Sprintf("Bad configuration: %s is not a %s.", path, file.Description()),
	}
}

//Error
func (err *ConfigError) Error() string {
	return err.msg
}
