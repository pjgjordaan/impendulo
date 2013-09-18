package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	errLogger, infoLogger *Logger
)

type (
	//Logger allows for concurrent logging.
	Logger struct {
		logger        *log.Logger
		console, file bool
	}
)

//init sets up the loggers.
func init() {
	logDir := filepath.Join(BaseDir(), "logs")
	y, m, d := time.Now().Date()
	dir := filepath.Join(logDir, strconv.Itoa(y), m.String(), strconv.Itoa(d))
	err := os.MkdirAll(dir, DPERM)
	if err != nil {
		panic(err)
	}
	errLogger, err = NewLogger(filepath.Join(dir, "E_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
	infoLogger, err = NewLogger(filepath.Join(dir, "I_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
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
