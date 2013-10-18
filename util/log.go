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

package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type (
	//Logger allows for concurrent logging.
	Logger struct {
		logger        *log.Logger
		console, file bool
	}
)

var (
	errLogger, infoLogger *Logger
	logDir                string
)

//init sets up the loggers.
func init() {
	y, m, d := time.Now().Date()
	dir := filepath.Join(LogDir(), strconv.Itoa(y), m.String(), strconv.Itoa(d))
	err := os.MkdirAll(dir, DPERM)
	if err != nil {
		panic(err)
	}
	now := time.Now().Format(layout)
	errLogger, err = NewLogger(filepath.Join(dir, "E_"+now+".log"))
	if err != nil {
		panic(err)
	}
	infoLogger, err = NewLogger(filepath.Join(dir, "I_"+now+".log"))
	if err != nil {
		panic(err)
	}
}

func LogDir() string {
	if logDir == "" {
		logDir = filepath.Join(BaseDir(), "logs")
	}
	return logDir
}

//SetErrorConsoleLogging sets whether errors should be logged to the console.
func SetErrorLogging(setting string) {
	errLogger.setLogging(setting)
}

//SetInfoConsoleLogging sets whether info should be logged to the console.
func SetInfoLogging(setting string) {
	infoLogger.setLogging(setting)
}

func (this *Logger) setLogging(setting string) {
	if setting == "a" {
		this.file = true
		this.console = true
	} else if setting == "f" {
		this.file = true
	} else if setting == "c" {
		this.console = true
	}
}

//Log safely logs data to this logger's log file.
func (this *Logger) Log(vals ...interface{}) {
	if this.file {
		this.logger.Print(vals)
	}
	if this.console {
		fmt.Println(vals)
	}
}

//NewLogger creates a new SyncLogger which writes its logs to the specified file.
func NewLogger(fname string) (logger *Logger, err error) {
	logFile, err := os.Create(fname)
	if err != nil {
		return
	}
	logger = &Logger{log.New(logFile, "", log.LstdFlags), false, false}
	return
}

//Log sends data to be logged to the appropriate logger.
func Log(v ...interface{}) {
	if len(v) > 0 {
		if _, ok := v[0].(error); ok {
			errLogger.Log(v)
		} else {
			infoLogger.Log(v)
		}
	}
}
