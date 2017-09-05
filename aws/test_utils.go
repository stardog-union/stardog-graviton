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
	"path"
)


func MakeTestSSH(rc int)  (string, string, error) {
	exedir, err := ioutil.TempDir("/tmp", "stardogtest")
	if err != nil {
		return "", "", err
	}

	fakeSSH := `#!/usr/bin/env bash
	exit %d`
	fakeSSH = fmt.Sprintf(fakeSSH, rc)

	sshFile := path.Join(exedir, "ssh")
	err = ioutil.WriteFile(sshFile, []byte(fakeSSH), 0755)
	if err != nil {
		return "", "", err
	}
	return exedir, sshFile, nil
}

func MakeTestTerraform(rc int, output string, dir string) (string, string, error) {
	var err error
	exedir := dir
	if dir == "" {
		exedir, err = ioutil.TempDir("/tmp", "stardogtest")
		if err != nil {
			return "", "", err
		}
	}
	paramsFile := path.Join(exedir, "params")
	dataFile := path.Join(exedir, "datafileTerraform")
	err = ioutil.WriteFile(dataFile, []byte(output), 0644)
	if err != nil {
		return "", "", fmt.Errorf("Failed to write the file %s", err)
	}

	fakeTerraform := `#!/usr/bin/env bash
	if [ "X$1" == "X--version" ]; then
		echo Terraform v%s
		exit 0
	fi
	echo ${@} > %s
	cat %s
	exit %d`
	fakeTerraform = fmt.Sprintf(fakeTerraform, TerraformVersion, paramsFile, dataFile, rc)
	terraformFile := path.Join(exedir, "terraform")
	err = ioutil.WriteFile(terraformFile, []byte(fakeTerraform), 0755)
	if err != nil {
		return "", "", err
	}
	return exedir, terraformFile, nil
}

func MakeTestPacker(rc int, output string, dir string) (string, string, error) {
	var err error
	exedir := dir
	if dir == "" {
		exedir, err = ioutil.TempDir("/tmp", "stardogtest")
		if err != nil {
			return "", "", err
		}
	}
	paramsFile := path.Join(exedir, "params")
	dataFile := path.Join(exedir, "datafilePacker")
	err = ioutil.WriteFile(dataFile, []byte(output), 0644)
	if err != nil {
		return "", "", fmt.Errorf("Failed to write the file %s", err)
	}

	fakePacker := `#!/usr/bin/env bash
	if [ "X$1" == "X--version" ]; then
		echo %s
		exit 0
	fi
	echo ${@} > %s
	cat %s
	exit %d`
	fakePacker = fmt.Sprintf(fakePacker, PackerVersion, paramsFile, dataFile, rc)
	packerFile := path.Join(exedir, "packer")
	err = ioutil.WriteFile(packerFile, []byte(fakePacker), 0755)
	if err != nil {
		return "", "", err
	}
	return exedir, packerFile, nil
}
