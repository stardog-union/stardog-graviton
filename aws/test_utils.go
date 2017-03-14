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

package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// CreateTestExec writes a file to the system for use as a mock for packer or terraform
// in testing.
func CreateTestExec(pgmName string, output string, rc int) (string, string, error) {
	exedir, err := ioutil.TempDir("/tmp", "stardogtest")
	if err != nil {
		return "", "", err
	}

	dataFile := path.Join(exedir, "datafile")
	err = ioutil.WriteFile(dataFile, []byte(output), 0644)
	if err != nil {
		return "", "", fmt.Errorf("Failed to write the file %s", err)
	}
	paramsFile := path.Join(exedir, "params")

	exeTemplate := fmt.Sprintf("#!/usr/bin/env bash\necho ${@} > %s\ncat %s\nexit %d", paramsFile, dataFile, rc)
	exeFile := path.Join(exedir, pgmName)
	err = ioutil.WriteFile(exeFile, []byte(exeTemplate), 0755)
	if err != nil {
		return "", "", fmt.Errorf("Failed to write the file %s", err)
	}
	startPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", exedir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		return "", "", fmt.Errorf("Failed to set env %s", err)
	}
	return exedir, exeFile, nil
}
