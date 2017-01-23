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
	"testing"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

func TestDeploymentLoadDefaults(t *testing.T) {

	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
	}
	i := make(map[string]string)
	i["region"] = "us-west-2"
	i["ami_id"] = "ami-deadbeef"
	i["aws_key_name"] = "somekey"
	i["zk_instance_type"] = "zkinst"
	i["sd_instance_type"] = "sdinst"

	err := plugin.LoadDefaults(&i)
	if err != nil {
		t.Fatalf("Failed to load defaults %s", err)
	}
	if plugin.Region != i["region"] {
		t.Fatalf("Region not set right")
	}
	if plugin.AmiID != i["ami_id"] {
		t.Fatalf("ami_id not set right")
	}
	if plugin.AwsKeyName != i["aws_key_name"] {
		t.Fatalf("aws_key_name not set right")
	}
	if plugin.ZkInstanceType != i["zk_instance_type"] {
		t.Fatalf("zk_instance_type not set right")
	}
	if plugin.SdInstanceType != i["sd_instance_type"] {
		t.Fatalf("sd_instance_type not set right")
	}
}

func TestDeploymentLoadEnvs(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
	}
	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	keySave := os.Getenv("AWS_ACCESS_KEY_ID")
	if keySave != "" {
		defer os.Setenv("AWS_ACCESS_KEY_ID", keySave)
	}
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	secretSave := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretSave != "" {
		defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretSave)
	}
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	baseD := sdutils.BaseDeployment{
		Type:       plugin.GetName(),
		Name:       "testdep",
		Directory:  dir,
		Version:    "4.2",
		PrivateKey: "/some/path",
	}
	_, err := plugin.DeploymentLoader(&app, &baseD, true)
	if err == nil {
		t.Fatalf("The deployment should have failed")
	}
	os.Setenv("AWS_SECRET_ACCESS_KEY", keySave)
	_, err = plugin.DeploymentLoader(&app, &baseD, true)
	if err == nil {
		t.Fatalf("The deployment should have failed")
	}
}

func TestDeploymentLoadNoExes(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
	}
	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	keySave := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", keySave)
	os.Setenv("AWS_ACCESS_KEY_ID", "somevalue")
	secretSave := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretSave)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somesecret")

	sPath := os.Getenv("PATH")
	defer os.Setenv("PATH", sPath)
	os.Setenv("PATH", "/")

	baseD := sdutils.BaseDeployment{
		Type:       plugin.GetName(),
		Name:       "testdep",
		Directory:  dir,
		Version:    "4.2",
		PrivateKey: "/some/path",
	}
	_, err := plugin.DeploymentLoader(&app, &baseD, true)
	if err == nil {
		t.Fatalf("The deployment should have failed")
	}
}

func TestDeploymentLoadNew(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtest")
	defer os.RemoveAll(dir)

	plugin := &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "notreal",
		AwsKeyName:     "somekey",
		ZkInstanceType: "m3.large",
		SdInstanceType: "m3.large",
	}
	app := sdutils.TestContext{
		ConfigDir: dir,
		Version:   "4.2",
	}

	keySave := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", keySave)
	os.Setenv("AWS_ACCESS_KEY_ID", "somevalue")
	secretSave := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretSave)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somesecret")

	exedirT, _, err := CreateTestExec("terraform", "data", 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedirT)
	exedirP, _, err := CreateTestExec("packer", "data", 0)
	if err != nil {
		t.Fatalf("Failed to write the file %s", err)
	}
	defer os.RemoveAll(exedirP)

	sPath := os.Getenv("PATH")
	defer os.Setenv("PATH", sPath)
	os.Setenv("PATH", fmt.Sprintf("%s:%s:%s", exedirP, exedirT, sPath))

	baseD := sdutils.BaseDeployment{
		Type:       plugin.GetName(),
		Name:       "testdep",
		Directory:  dir,
		Version:    "4.2",
		PrivateKey: "/some/path",
	}
	_, err = plugin.DeploymentLoader(&app, &baseD, true)
	if err != nil {
		t.Fatalf("The deployment should not have failed %s", err)
	}
	_, err = plugin.DeploymentLoader(&app, &baseD, false)
	if err != nil {
		t.Fatalf("The deployment should not have failed %s", err)
	}
}
