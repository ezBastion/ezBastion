// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

// Package logmanager add helper for logrus
package logmanager

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	ezbevent "ezBastion/pkg/eventlogmanager"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type callInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

// EventLog management
var level string

// SetLogLevel set logrus level
func SetLogLevel(LogLevel string, exPath string, fileName string, maxSize int, maxBackups int, maxAge int, IsWindowsService bool) error {
	log.SetFormatter(&log.JSONFormatter{})
	level = LogLevel
	switch LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "critical":
		log.SetLevel(log.FatalLevel)
	default:
		fmt.Errorf("logmanager/SetLogLevel() failed: Bad log level name, set to Info")
		log.SetLevel(log.InfoLevel)
	}

	// Adding the method and line caller, easier to debug

	log.SetReportCaller(!IsWindowsService)

	abspathfilename := exPath + string(os.PathSeparator) + fileName
	lj := &lumberjack.Logger{
		Filename:   abspathfilename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}

	if IsWindowsService {
		log.SetOutput(lj)
	} else {
		mWriter := io.MultiWriter(os.Stderr, lj)
		log.SetOutput(mWriter)
	}
	log.Debug("Log system initialized.")

	return nil

}

func retrieveCallInfo() *callInfo {
	pc, file, line, _ := runtime.Caller(2)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return &callInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

func WithFields(s1 string, s2 string) {
	log.WithFields(log.Fields{s1: s2})
}

func StartWindowsEvent(name string) {
	if ezbevent.Status == 0 {
		ezbevent.Open(name)
	}

}

// Info logs an info event into the windows eventlog system
func Debug(logline string) error {
	log.Debugln(logline)
	if level == "debug" {
		if ezbevent.Status == 0 {
			ezbevent.Elog.Info(1, "DEBUG : "+logline)
		}
	}
	return nil
}

func Info(logline string, forceStdout ...bool) error {
	log.Infoln(logline)
	output := false
	if len(forceStdout) > 0 {
		output = forceStdout[0]
	}

	if output {
		fmt.Println(logline)
	}
	if level == "debug" || (level == "info") {
		if ezbevent.Status == 0 {
			ezbevent.Elog.Info(1, logline)
		}
	}
	return nil
}

// Error logs an error event into the windows eventlog system
func Error(logline string) error {
	log.Errorln(logline)
	if (level == "info") || (level == "warning") || (level == "error") || (level == "debug") {
		if ezbevent.Status == 0 {
			ezbevent.Elog.Error(1, logline)
		}

	}
	return nil
}

// Warning logs an warning event into the windows eventlog system
func Warning(logline string) error {
	log.Warnln(logline)
	if (level == "debug") || (level == "info") || (level == "warning") {
		if ezbevent.Status == 0 {
			ezbevent.Elog.Warning(1, logline)
		}
	}
	return nil
}

func Fatal(logline string) {
	log.Fatal(logline)
}
