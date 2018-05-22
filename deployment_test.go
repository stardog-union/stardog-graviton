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
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestGetNoPluging(t *testing.T) {
	_, err := GetPlugin("not real")
	if err == nil {
		t.Fatal("No pluging should exist")
	}
}

func TestLoadDeploymentNoPlugin(t *testing.T) {
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		t.Fatal("Temp dir failed")
	}
	defer os.RemoveAll(dir)

	app := TestContext{
		ConfigDir: dir,
		Version:   "test1",
	}
	baseD := BaseDeployment{
		Type: "notreal",
		Name: "notreal",
	}

	_, err = LoadDeployment(&app, &baseD, false)
	if err == nil {
		t.Fatal("No pluging should exist")
	}
}

func TestLoadDeploymentNoLoadPlugin(t *testing.T) {
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		t.Fatal("Temp dir failed")
	}
	defer os.RemoveAll(dir)

	app := TestContext{
		ConfigDir: dir,
		Version:   "test1",
	}
	plugin := &tstPlugin{}
	baseD := BaseDeployment{
		Type: plugin.GetName(),
		Name: "notreal",
	}
	AddCloudType(plugin)

	_, err = LoadDeployment(&app, &baseD, false)
	if err == nil {
		t.Fatal("No deployment load should fail")
	}
}

func TestLoadNewDeploymentFailLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		t.Fatal("Temp dir failed")
	}
	defer os.RemoveAll(dir)

	app := TestContext{
		ConfigDir: dir,
		Version:   "test1",
	}
	plugin := &tstPlugin{}
	baseD := BaseDeployment{
		Type:      plugin.GetName(),
		Name:      "notreal",
		Directory: dir,
	}
	AddCloudType(plugin)

	_, err = LoadDeployment(&app, &baseD, true)
	if err == nil {
		t.Fatal("No deployment load should fail")
	}
}

func TestLoadNewDeploymentGoodLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		t.Fatal("Temp dir failed")
	}
	defer os.RemoveAll(dir)

	app := TestContext{
		ConfigDir: dir,
		Version:   "test1",
	}
	plugin := &tstPlugin{Dep: &tpDeployment{}}
	baseD := BaseDeployment{
		Type: plugin.GetName(),
		Name: "notreal",
	}
	AddCloudType(plugin)

	_, err = LoadDeployment(&app, &baseD, true)
	if err != nil {
		t.Fatalf("The deployment should load %s", err)
	}
	if !PathExists(dir) {
		t.Fatal("The deployment directory should not be created when failure occurs.")
	}
	DeleteDeployment(&app, baseD.Name)
}

func TestDeployDir(t *testing.T) {
	depDir := DeploymentDir("/a/base/dir", "depname")
	if depDir != "/a/base/dir/deployments/depname" {
		t.Fatal("The dir is not what we expected")
	}
}

func TestLoadExistingDeploymentGoodLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		t.Fatal("Temp dir failed")
	}
	defer os.RemoveAll(dir)

	app := TestContext{
		ConfigDir: dir,
		Version:   "test1",
	}
	plugin := &tstPlugin{Dep: &tpDeployment{}}
	baseD := BaseDeployment{
		Type:      plugin.GetName(),
		Name:      "notreal",
		Directory: dir,
	}
	AddCloudType(plugin)

	confPath := path.Join(baseD.Directory, "config.json")
	WriteJSON(&baseD, confPath)

	_, err = LoadDeployment(&app, &baseD, true)
	if err != nil {
		t.Fatalf("The deployment should load %s", err)
	}
	if !PathExists(dir) {
		t.Fatal("The deployment directory should not be created when failure occurs.")
	}
	DeleteDeployment(&app, baseD.Name)
}
