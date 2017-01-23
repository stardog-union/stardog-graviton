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
	"testing"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

func TestGoodPacker(t *testing.T) {
	fakePacker := "#!/usr/bin/env bash\necho amazon-ebs,artifact,0,string,AMIs were created:ami-deadbeef\nexit 0"

	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	packerFile := path.Join(dir, "packer")
	err := ioutil.WriteFile(packerFile, []byte(fakePacker), 0755)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)

	newPath := fmt.Sprintf("%s:%s", dir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set env %s", err)
	}

	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	awsP := GetPlugin()
	err = awsP.BuildImage(&app, "/etc/group", "4.2")
	if err != nil {
		t.Fatalf("Packer failed %s", err)
	}
	if !awsP.HaveImage(&app) {
		t.Fatalf("The image should be there")
	}
}

func TestBadRcPacker(t *testing.T) {
	fakePacker := "#!/usr/bin/env bash\necho amazon-ebs,artifact,0,string,AMIs were created:ami-deadbeef\nexit 1"

	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	packerFile := path.Join(dir, "packer")
	err := ioutil.WriteFile(packerFile, []byte(fakePacker), 0755)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)

	newPath := fmt.Sprintf("%s:%s", dir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set env %s", err)
	}

	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	awsP := GetPlugin()
	err = awsP.BuildImage(&app, "/etc/group", "4.2")
	if err == nil {
		t.Fatalf("Should have failed")
	}
	if awsP.HaveImage(&app) {
		t.Fatalf("The image should not be there")
	}
}

func TestBadAMIPacker(t *testing.T) {
	fakePacker := "#!/usr/bin/env bash\necho amazon-ebs,artifact,0,string,AMIs were created:NOAMI\nexit 0"

	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	packerFile := path.Join(dir, "packer")
	err := ioutil.WriteFile(packerFile, []byte(fakePacker), 0755)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)

	newPath := fmt.Sprintf("%s:%s", dir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set env %s", err)
	}

	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	awsP := GetPlugin()
	err = awsP.BuildImage(&app, "/etc/group", "4.2")
	if err == nil {
		t.Fatalf("Should have failed")
	}
	if awsP.HaveImage(&app) {
		t.Fatalf("The image should not be there")
	}
}

func TestHaveNoAmi(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}
	awsP := GetPlugin()
	if awsP.HaveImage(&app) {
		t.Fatalf("The image should not be there")
	}
}
