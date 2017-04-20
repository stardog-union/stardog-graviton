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
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

func TestInstanceNotThere(t *testing.T) {
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

	inst, err := NewEc2Instance(&app, dd)
	if inst.InstanceExists() {
		t.Fatalf("The instance should not exist")
	}
	err = inst.DeleteInstance()
	if err == nil {
		t.Fatalf("The instance should not exist for deletion")
	}
	err = inst.Status()
	if err == nil {
		t.Fatalf("The instance should not exist for status")
	}
	err = inst.CreateInstance(8, 1, 60)
	if err == nil {
		t.Fatalf("The instance should not exist for client")
	}
	err = inst.OpenInstance(8, 1, "0.0.0.0/0", 60)
	if err == nil {
		t.Fatalf("The instance should not exist for client")
	}
}

func TestInstanceFakeTerraform(t *testing.T) {
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

	inst, err := NewEc2Instance(&app, dd)
	if inst.InstanceExists() {
		t.Fatalf("The instance should not exist")
	}

	data := `{
    "bastion_contact": {
        "sensitive": false,
        "type": "string",
        "value": "XXX-1392906412.us-west-1.elb.amazonaws.com"
    },
    "stardog_contact": {
        "sensitive": false,
        "type": "string",
        "value": "XXX-1603084014.us-west-1.elb.amazonaws.com"
    },
    "stardog_internal_contact": {
        "sensitive": false,
        "type": "string",
        "value": "XXX-160332144.us-west-1.elb.amazonaws.com"
    },
    "zookeeper_nodes": {
        "sensitive": false,
        "type": "list",
        "value": [
            "XXXb0-1728919283.us-west-1.elb.amazonaws.com",
            "XXXzkelb1-8734281.us-west-1.elb.amazonaws.com",
            "XXXg2zkelb2-504548219.us-west-1.elb.amazonaws.com"
        ]
    }
}`

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exedir, _, err := CreateTestExec("terraform", data, 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir)

	ebs := NewAwsEbsVolumeManager(&app, dd)
	err = ebs.CreateSet("/path/", 1, 3)
	if err != nil {
		t.Fatalf("The create should have worked")
	}

	err = inst.CreateInstance(8, 1, 60)
	if err != nil {
		t.Fatalf("The instance should exist for client %s", err)
	}
	err = inst.OpenInstance(8, 1, "0.0.0.0/0", 60)
	if err != nil {
		t.Fatalf("The instance should exist for client %s", err)
	}
	err = inst.Status()
	if err != nil {
		t.Fatalf("The instance should exist for status %s", err)
	}
	err = inst.DeleteInstance()
	if err != nil {
		t.Fatalf("The instance should exist for deletion %s", err)
	}
}

func TestInstanceFakeTerraformThroughDeployment(t *testing.T) {
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

	data := `{
    "bastion_contact": {
        "sensitive": false,
        "type": "string",
        "value": "bastion.com"
    },
    "stardog_contact": {
        "sensitive": false,
        "type": "string",
        "value": "YYY-1603084014.us-west-1.elb.amazonaws.com"
    },
    "stardog_internal_contact": {
        "sensitive": false,
        "type": "string",
        "value": "XXX-160332144.us-west-1.elb.amazonaws.com"
    },
    "zookeeper_nodes": {
        "sensitive": false,
        "type": "list",
        "value": [
            "YYYb0-1728919283.us-west-1.elb.amazonaws.com",
            "YYYzkelb1-8734281.us-west-1.elb.amazonaws.com",
            "YYYg2zkelb2-504548219.us-west-1.elb.amazonaws.com"
        ]
    }
}`

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exedir, _, err := CreateTestExec("terraform", data, 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedir)

	ebs := NewAwsEbsVolumeManager(&app, dd)
	err = ebs.CreateSet("/path/", 1, 3)
	if err != nil {
		t.Fatalf("The create should have worked")
	}

	err = dd.CreateInstance(8, 1, 60)
	if err != nil {
		t.Fatalf("The instance should exist for client %s", err)
	}
	err = dd.OpenInstance(8, 1, "0.0.0.0/0", 60)
	if err != nil {
		t.Fatalf("The instance should exist for client %s", err)
	}
	err = dd.StatusInstance()
	if err != nil {
		t.Fatalf("The instance should exist for status %s", err)
	}
	if !dd.InstanceExists() {
		t.Fatalf("The instance should exist for status %s", err)
	}
	sd, err := dd.FullStatus()
	if err != nil {
		t.Fatalf("The instance should exist for deletion %s", err)
	}
	if sd.SSHHost != "bastion.com" {
		t.Fatalf("The bastion host was not correct")
	}
	err = dd.DeleteInstance()
	if err != nil {
		t.Fatalf("The instance should exist for deletion %s", err)
	}
}

func TestInstanceNotThereThroughDd(t *testing.T) {
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

	err = dd.DeleteInstance()
	if err == nil {
		t.Fatalf("The instance should not exist for deletion")
	}
	err = dd.StatusInstance()
	if err == nil {
		t.Fatalf("The instance should not exist for status")
	}
	err = dd.CreateInstance(8, 1, 60)
	if err == nil {
		t.Fatalf("The instance should not exist for client")
	}
}
