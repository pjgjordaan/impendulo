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
	dir, e := LogDir()
	if e != nil {
		panic(e)
	}
	dir = filepath.Join(dir, strconv.Itoa(y), m.String(), strconv.Itoa(d))
	if e = os.MkdirAll(dir, DPERM); e != nil {
		panic(e)
	}
	n := time.Now().Format(layout)
	errLogger, e = NewLogger(filepath.Join(dir, "E_"+n+".log"))
	if e != nil {
		panic(e)
	}
	infoLogger, e = NewLogger(filepath.Join(dir, "I_"+n+".log"))
	if e != nil {
		panic(e)
	}
}

func LogDir() (string, error) {
	if logDir != "" {
		return logDir, nil
	}
	b, e := BaseDir()
	if e != nil {
		return "", e
	}
	logDir = filepath.Join(b, "logs")
	return logDir, nil
}

//SetErrorConsoleLogging sets whether errors should be logged to the console.
func SetErrorLogging(s string) {
	errLogger.setLogging(s)
}

//SetInfoConsoleLogging sets whether info should be logged to the console.
func SetInfoLogging(s string) {
	infoLogger.setLogging(s)
}

func (l *Logger) setLogging(s string) {
	switch s {
	case "a":
		l.file = true
		l.console = true
	case "f":
		l.file = true
	case "c":
		l.console = true
	}
}

//Log safely logs data to this logger's log file.
func (l *Logger) Log(v ...interface{}) {
	if l.file {
		l.logger.Print(v)
	}
	if l.console {
		fmt.Println(v)
	}
}

//NewLogger creates a new SyncLogger which writes its logs to the specified file.
func NewLogger(n string) (*Logger, error) {
	f, e := os.Create(n)
	if e != nil {
		return nil, e
	}
	return &Logger{log.New(f, "", log.LstdFlags), false, false}, nil
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
