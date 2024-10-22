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
	"errors"
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
		Bin     map[Bin]string
		Cfg     map[Cfg]string
		Dir     map[Dir]string
		Jar     map[Jar]string
		Sh      map[Sh]string
		Archive map[Archive]string
	}
	//Bin is a string type for paths to binary files.
	Bin string
	//Cfg is a string type for paths to configuration files.
	Cfg string
	//Dir is a string type for paths to directories.
	Dir     string
	Archive string
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
	GCC   Bin = "gcc"
	MAKE  Bin = "make"

	//Configurations
	CHECKSTYLE_CFG Cfg = "checkstyle_cfg"
	PMD_CFG        Cfg = "pmd_cfg"

	//Directories
	JPF_FINDER    Dir = "jpf_finder"
	JPF_HOME      Dir = "jpf_home"
	JPF_RUNNER    Dir = "jpf_runner"
	JUNIT_TESTING Dir = "junit_testing"
	JACOCO_HOME   Dir = "jacoco_home"

	//Archives
	INTLOLA Archive = "intlola"

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
	config           *Configuration
	defaultFile      string
	UnitialisedError = errors.New("configuration not initialised")
)

func init() {
	l, e := DefaultConfig()
	if e == nil {
		e = LoadConfigs(l)
	}
	if e != nil {
		util.Log(e)
	}
}

//DefaultConfig retrieves the default configuration file path.
//This is $IMPENDULO_PATH/config/config.json
func DefaultConfig() (string, error) {
	if defaultFile != "" {
		return defaultFile, nil
	}
	b, e := util.BaseDir()
	if e != nil {
		return "", e
	}
	defaultFile = filepath.Join(b, "config.json")
	if util.IsFile(defaultFile) {
		return defaultFile, nil
	} else if util.IsDir(defaultFile) {
		return "", fmt.Errorf("%s is a directory", defaultFile)
	}
	ip, e := util.InstallPath()
	if e != nil {
		return "", e
	}
	return defaultFile, util.CopyFile(defaultFile, filepath.Join(ip, "config", "config.json"))
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
func LoadConfigs(fname string) error {
	//Load configuration from Json file.
	cfgFile, e := os.Open(fname)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(util.ReadBytes(cfgFile), &config); e != nil {
		return e
	}
	ip, e := util.InstallPath()
	if e != nil {
		return e
	}
	jp := filepath.Join(ip, "java")
	config.Dir[JPF_RUNNER] = filepath.Join(jp, "runner")
	config.Dir[JPF_FINDER] = filepath.Join(jp, "finder")
	config.Dir[JUNIT_TESTING] = filepath.Join(jp, "testing")
	//Check if configurations are valid.
	for b, p := range config.Bin {
		if e = b.Valid(p); e != nil {
			return e
		}
	}
	for c, p := range config.Cfg {
		if e = c.Valid(p); e != nil {
			return e
		}
	}
	for d, p := range config.Dir {
		if e = d.Valid(p); e != nil {
			return e
		}
		if d == JPF_HOME {
			jar := filepath.Join(p, filepath.Join("build", "jpf.jar"))
			if e = JPF.Valid(jar); e != nil {
				return e
			}
			config.Jar[JPF] = jar
			jar = filepath.Join(p, filepath.Join("build", "RunJPF.jar"))
			if e = JPF_RUN.Valid(jar); e != nil {
				return e
			}
			config.Jar[JPF_RUN] = jar
		}
	}
	for j, p := range config.Jar {
		if e = j.Valid(p); e != nil {
			return e
		}
	}
	for sh, p := range config.Sh {
		if e = sh.Valid(p); e != nil {
			return e
		}
	}
	for a, p := range config.Archive {
		if e = a.Valid(p); e != nil {
			return e
		}
	}
	return nil
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

func (a Archive) Valid(path string) error {
	return valid(a, path, util.IsFile)
}

func (a Archive) Description() string {
	return "archive"
}

func (a Archive) Path() (string, error) {
	return path(a)
}

//valid determines whether the provided path is valid for corresponding file.
func valid(f File, p string, v Validator) error {
	if !v(p) {
		return Invalid(f, p)
	}
	return nil
}

func (c *Configuration) Path(f File) (string, error) {
	var ok bool
	var p string
	switch t := f.(type) {
	case Bin:
		p, ok = c.Bin[t]
	case Cfg:
		p, ok = c.Cfg[t]
	case Dir:
		p, ok = c.Dir[t]
	case Jar:
		p, ok = c.Jar[t]
	case Sh:
		p, ok = c.Sh[t]
	case Archive:
		p, ok = c.Archive[t]
	default:
		return "", &ConfigError{msg: fmt.Sprintf("Unknown configuration type %s.", t)}
	}
	if !ok {
		return "", NA(f)
	}
	return p, nil
}

//path attempts to retrieve the file's path.
func path(f File) (string, error) {
	if config == nil {
		return "", UnitialisedError
	}
	return config.Path(f)
}

//NA creates a new ConfigError if a request is made for an unavailable configuration.
func NA(f File) error {
	return &ConfigError{msg: fmt.Sprintf("Config %s not found.", f)}
}

//Invalid creates a new ConfigError for a bad configuration specification.
func Invalid(f File, p string) error {
	return &ConfigError{msg: fmt.Sprintf("Bad configuration: %s is not a %s.", p, f.Description())}
}

//Error
func (err *ConfigError) Error() string {
	return err.msg
}
