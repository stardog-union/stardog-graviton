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
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestWarnLevel(t *testing.T) {
	buf := bytes.NewBufferString("")
	realLogger := log.New(buf, "", log.Ldate|log.Ltime)
	sdLog, err := NewSdVaLogger(realLogger, "WARN")
	if err != nil {
		t.Fatalf("Failed to load the logger %s", err)
	}
	msg := "Just Some message"
	sdLog.Logf(WARN, msg)
	sdLog.Logf(ERROR, msg)
	sdLog.Logf(DEBUG, msg)

	if !strings.Contains(buf.String(), "WARN") {
		t.Fatalf("A warning should have been found")
	}
	if !strings.Contains(buf.String(), "ERROR") {
		t.Fatalf("A error should have been found")
	}
	if strings.Contains(buf.String(), "DEBUG") {
		t.Fatalf("Debug should not have been found")
	}
}

func TestErrorLevel(t *testing.T) {
	buf := bytes.NewBufferString("")
	realLogger := log.New(buf, "", log.Ldate|log.Ltime)
	sdLog, err := NewSdVaLogger(realLogger, "ERROR")
	if err != nil {
		t.Fatalf("Failed to load the logger %s", err)
	}
	msg := "Just Some message"
	sdLog.Logf(WARN, msg)
	sdLog.Logf(ERROR, msg)
	sdLog.Logf(INFO, msg)
	sdLog.Logf(DEBUG, msg)

	if !strings.Contains(buf.String(), "ERROR") {
		t.Fatalf("A error should have been found")
	}
	if strings.Contains(buf.String(), "WARN") {
		t.Fatalf("A warning should not have been found")
	}
	if strings.Contains(buf.String(), "INFO") {
		t.Fatalf("A info should not have been found")
	}
	if strings.Contains(buf.String(), "DEBUG") {
		t.Fatalf("Debug should not have been found")
	}
}

func TestInfoLevel(t *testing.T) {
	buf := bytes.NewBufferString("")
	realLogger := log.New(buf, "", log.Ldate|log.Ltime)
	sdLog, err := NewSdVaLogger(realLogger, "INFO")
	if err != nil {
		t.Fatalf("Failed to load the logger %s", err)
	}
	msg := "Just Some message"
	sdLog.Logf(WARN, msg)
	sdLog.Logf(ERROR, msg)
	sdLog.Logf(DEBUG, msg)
	sdLog.Logf(INFO, msg)

	if !strings.Contains(buf.String(), "WARN") {
		t.Fatalf("A warn should have been found")
	}
	if !strings.Contains(buf.String(), "ERROR") {
		t.Fatalf("A error should have been found")
	}
	if strings.Contains(buf.String(), "DEBUG") {
		t.Fatalf("Debug should not have been found")
	}
}

func TestDebugLevel(t *testing.T) {
	buf := bytes.NewBufferString("")
	realLogger := log.New(buf, "", log.Ldate|log.Ltime)
	sdLog, err := NewSdVaLogger(realLogger, "DEBUG")
	if err != nil {
		t.Fatalf("Failed to load the logger %s", err)
	}
	msg := "Just Some message"
	sdLog.Logf(WARN, msg)
	sdLog.Logf(ERROR, msg)
	sdLog.Logf(DEBUG, msg)
	sdLog.Logf(INFO, msg)

	if !strings.Contains(buf.String(), "WARN") {
		t.Fatalf("A warning should have been found")
	}
	if !strings.Contains(buf.String(), "ERROR") {
		t.Fatalf("A error should have been found")
	}
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Fatalf("A debug should have been found")
	}
	if !strings.Contains(buf.String(), "INFO") {
		t.Fatalf("A info should have been found")
	}
}

func TestInvalidLevel(t *testing.T) {
	buf := bytes.NewBufferString("")
	realLogger := log.New(buf, "", log.Ldate|log.Ltime)
	_, err := NewSdVaLogger(realLogger, "NOGOOD")
	if err == nil {
		t.Fatalf("The logger should not have loaded")
	}
}
