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

func TestVolumesNotThere(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)
	sshKeyFile := path.Join(dir, "keyfile")
	fakeOut := "xxx"
	ioutil.WriteFile(sshKeyFile, []byte(fakeOut), 0600)
	keySave := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", keySave)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	secretSave := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretSave)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somesecret")

	version := "4.2"
	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   version,
	}
	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
		VolumeType:     "gp2",
	}
	baseD := sdutils.BaseDeployment{
		Type:       plugin.GetName(),
		Name:       "testdep",
		Directory:  dir,
		Version:    version,
		PrivateKey: sshKeyFile,
	}
	dd, err := newAwsDeploymentDescription(&app, &baseD, plugin)
	if err != nil {
		t.Fatalf("Failed to make the deployment manager %s", err)
	}
	ebs := NewAwsEbsVolumeManager(&app, dd)
	if ebs.VolumeExists() {
		t.Fatalf("The volume shouldn't exist yet")
	}
	err = ebs.Status()
	if err == nil {
		t.Fatalf("The volume shouldn't exist yet, status should fail")
	}
	err = ebs.DeleteSet()
	if err == nil {
		t.Fatalf("The delete should have failed")
	}
	err = ebs.CreateSet("/path/", 1, 3)
	if err == nil {
		t.Fatalf("The create should have failed")
	}

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exedir, _, err := CreateTestExec("terraform", "data", 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir)

	newPath := fmt.Sprintf("%s:%s", exedir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set env %s", err)
	}

	err = ebs.CreateSet("/path/", 1, 3)
	if err != nil {
		t.Fatalf("The create should have worked")
	}
	if !ebs.VolumeExists() {
		t.Fatalf("The volume should exist yet")
	}

	deployDir := sdutils.DeploymentDir(dir, baseD.Name)
	volumeDir := path.Join(deployDir, "etc", "terraform", "volumes")
	loadedEbs, err := LoadEbsVolume(&app, volumeDir)
	if err != nil {
		t.Fatalf("The re-load should not have failed %s", err)
	}
	if loadedEbs.VolumeExists() {
		t.Fatalf("The volume reload should exist yet")
	}

	err = ebs.Status()
	if err == nil {
		t.Fatalf("The status should have bad output")
	}
	// Add in good status
	data := `{"volumes": { "sensitive": false, "type": "list", "value": ["vol-46313ce8", "vol-a34ba11c", "vol-50313cfe"]}}`
	exedir2, _, err := CreateTestExec("terraform", data, 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir2)

	err = ebs.Status()
	if err != nil {
		t.Fatalf("The status should work %s", err)
	}

	err = ebs.DeleteSet()
	if err != nil {
		t.Fatalf("The delete should not have failed")
	}
}

func TestVolumesThroughDD(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)
	sshKeyFile := path.Join(dir, "keyfile")
	fakeOut := "xxx"
	ioutil.WriteFile(sshKeyFile, []byte(fakeOut), 0600)
	keySave := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", keySave)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	secretSave := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretSave)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somesecret")

	version := "4.2"
	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   version,
	}
	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
		VolumeType:     "gp2",
	}
	baseD := sdutils.BaseDeployment{
		Type:       plugin.GetName(),
		Name:       "testdep",
		Directory:  dir,
		Version:    version,
		PrivateKey: sshKeyFile,
	}
	dd, err := newAwsDeploymentDescription(&app, &baseD, plugin)
	if err != nil {
		t.Fatalf("Failed to make the deployment manager %s", err)
	}
	if dd.VolumeExists() {
		t.Fatalf("The volume shouldn't exist yet")
	}
	err = dd.StatusVolumeSet()
	if err == nil {
		t.Fatalf("The volume shouldn't exist yet, status should fail")
	}
	err = dd.DeleteVolumeSet()
	if err == nil {
		t.Fatalf("The delete should have failed")
	}
	err = dd.CreateVolumeSet("/path/", 1, 3)
	if err == nil {
		t.Fatalf("The create should have failed")
	}

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exedir, _, err := CreateTestExec("terraform", "data", 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir)

	newPath := fmt.Sprintf("%s:%s", exedir, startPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set env %s", err)
	}

	err = dd.CreateVolumeSet("/path/", 1, 3)
	if err != nil {
		t.Fatalf("The create should have worked")
	}
	if !dd.VolumeExists() {
		t.Fatalf("The volume should exist yet")
	}

	deployDir := sdutils.DeploymentDir(dir, baseD.Name)
	volumeDir := path.Join(deployDir, "etc", "terraform", "volumes")
	loadedEbs, err := LoadEbsVolume(&app, volumeDir)
	if err != nil {
		t.Fatalf("The re-load should not have failed %s", err)
	}
	if loadedEbs.VolumeExists() {
		t.Fatalf("The volume reload should exist yet")
	}

	err = dd.StatusVolumeSet()
	if err == nil {
		t.Fatalf("The status should have bad output")
	}
	// Add in good status
	data := `{"volumes": { "sensitive": false, "type": "list", "value": ["vol-46313ce8", "vol-a34ba11c", "vol-50313cfe"]}}`
	exedir2, _, err := CreateTestExec("terraform", data, 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir2)

	err = dd.StatusVolumeSet()
	if err != nil {
		t.Fatalf("The status should work %s", err)
	}

	err = dd.DeleteVolumeSet()
	if err != nil {
		t.Fatalf("The delete should not have failed")
	}
}
