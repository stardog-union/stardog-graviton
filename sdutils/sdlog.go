//
//  Copyright (c) 2017, Stardog Union. <http://stardog.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sdutils

import (
	"fmt"
	"log"
	"strings"
)

const (
	// ERROR will only log error level messages
	ERROR = 1
	// WARN will log ERROR and WARN level messages
	WARN = 2
	// INFO is the default level and will log information, warnings and errors
	INFO = 3
	// DEBUG is a very verbose log level and should be used only for finding problems
	DEBUG = 4
)

var (
	// This map is setup to quickly translate from log level to string at log time
	debugToStringMap = make(map[int]string)
	// LogLevelNames is an array of strings that define all the valid log levels
	LogLevelNames []string
)

func init() {
	debugToStringMap[DEBUG] = "DEBUG"
	debugToStringMap[INFO] = "INFO"
	debugToStringMap[WARN] = "WARN"
	debugToStringMap[ERROR] = "ERROR"

	LogLevelNames = make([]string, len(debugToStringMap), len(debugToStringMap))
	i := 0
	for _, v := range debugToStringMap {
		LogLevelNames[i] = v
		i = i + 1
	}
}

// SdVaLogger is the interface to the Stardog Logger
type SdVaLogger interface {
	Logf(level int, format string, v ...interface{})
}

type sdLogger struct {
	logLevel int
	logger   *log.Logger
}

// NewSdVaLogger creates a new Stardog logging object from a system logger
func NewSdVaLogger(realLogger *log.Logger, logLevel string) (SdVaLogger, error) {
	var logger sdLogger
	logLevel = strings.ToUpper(logLevel)

	switch logLevel {
	case "DEBUG":
		logger.logLevel = DEBUG
	case "INFO":
		logger.logLevel = INFO
	case "WARN":
		logger.logLevel = WARN
	case "ERROR":
		logger.logLevel = ERROR
	default:
		return nil, fmt.Errorf("The log level must be one of DEBUG, INFO, WARN, or ERROR")
	}
	logger.logger = realLogger

	return &logger, nil
}

func (l *sdLogger) logit(lineLevel int, format string, v ...interface{}) {
	if lineLevel > l.logLevel {
		return
	}
	format = fmt.Sprintf("[%s] ", debugToStringMap[lineLevel]) + format
	l.logger.Printf(format, v...)
}

func (l *sdLogger) Logf(level int, format string, v ...interface{}) {
	l.logit(level, format, v...)
}
