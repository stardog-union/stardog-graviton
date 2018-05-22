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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"path/filepath"

	"github.com/stardog-union/stardog-graviton/aws"
	"github.com/stardog-union/stardog-graviton"
	"errors"
)

var (
	goodoutput1 = `{
    "bastion_contact": {
        "sensitive": false,
        "type": "string",
        "value": "bastion.sometest.com"
    },
    "stardog_contact": {
        "sensitive": false,
        "type": "string",
        "value": "stardog.sometest.com"
    },
    "stardog_internal_contact": {
        "sensitive": false,
        "type": "string",
        "value": "stardoginternal.sometest.com"
    },
    "zookeeper_nodes": {
        "sensitive": false,
        "type": "list",
        "value": [
            "zk0.sometest.com",
            "zk1.sometest.com",
            "zk2.sometest.com"
        ]
    },
    "volumes": { "sensitive": false, "type": "list", "value": ["vol-46313ce8", "vol-a34ba11c", "vol-50313cfe"]}
}`
	packerFile    = filepath.Join(os.TempDir(), "packer")
	terraformFile = filepath.Join(os.TempDir(), "terraform")
)

func TestAbout(t *testing.T) {
	dir, _ := ioutil.TempDir("", "stardogtests")
	confDir := path.Join(dir, "a", "new", "dir")
	defer os.RemoveAll(confDir)

	rc := realMain([]string{"--config-dir", confDir, "about"})
	if rc != 0 {
		t.Fatal("About should return 0")
	}
	if !sdutils.PathExists(confDir) {
		t.Fatalf("The conf dir should have been created %s", confDir)
	}
}

func TestDeploymentList(t *testing.T) {
	confDir, _ := ioutil.TempDir("", "stardogtests")
	sshKeyFile := path.Join(confDir, "keyfile")
	fakeOut := "xxx"
	ioutil.WriteFile(sshKeyFile, []byte(fakeOut), 0600)

	defer os.RemoveAll(confDir)

	consoleLog := path.Join(confDir, "output0")
	depname1 := randDeployName()
	depname2 := randDeployName()

	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", awsKeyID)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	awsSecretKeyID := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecretKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somevalue")

	rc := realMain([]string{"--quiet", "--config-dir", confDir, "--console-file", consoleLog, "deployment", "list"})
	if rc != 0 {
		t.Fatal("About should return 0")
	}
	b, err := ioutil.ReadFile(consoleLog)
	if err != nil {
		t.Fatal("the console log was not created")
	}
	if strings.TrimSpace(string(b)) != "" {
		t.Fatal("There should have been no output")
	}

	err = buildImage("ami-beefwest", confDir, "4.2", "/etc/group", "us-west-1")
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}

	rc = realMain([]string{"--quiet", "--config-dir", confDir, "deployment", "new", "--private-key", sshKeyFile, "--aws-key-name", "keyname", depname1, "4.2"})
	if rc != 0 {
		t.Fatal("dep new should return 0")
	}
	rc = realMain([]string{"--quiet", "--config-dir", confDir, "deployment", "new", "--private-key", sshKeyFile, "--aws-key-name", "keyname", depname2, "4.2"})
	if rc != 0 {
		t.Fatal("dep new should return 0")
	}

	rc = realMain([]string{"--quiet", "--config-dir", confDir, "--console-file", consoleLog, "deployment", "list"})
	if rc != 0 {
		t.Fatal("About should return 0")
	}
	b, err = ioutil.ReadFile(consoleLog)
	if err != nil {
		t.Fatal("the console log was not created")
	}
	output := string(b)
	if !strings.Contains(output, depname1) {
		t.Fatalf("The deployment %s was not listed", depname1)
	}
	if !strings.Contains(output, depname2) {
		t.Fatalf("The deployment %s was not listed", depname2)
	}

	rc = realMain([]string{"--quiet", "--config-dir", confDir, "deployment", "destroy", "--force", depname1})
	if rc != 0 {
		t.Fatal("dep destroy should return 0")
	}
	rc = realMain([]string{"--quiet", "--config-dir", confDir, "deployment", "destroy", "--force", depname2})
	if rc != 0 {
		t.Fatal("dep destroy should return 0")
	}
}

func TestStatusNoDeploy(t *testing.T) {
	rc := realMain([]string{"--quiet", "status"})
	if rc == 0 {
		t.Fatal("Should have failed")
	}
}

func buildImage(amiName string, confDir string, version string, releasefile string, region string) error {
	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	data := fmt.Sprintf("amazon-ebs,artifact,0,string,AMIs were created:%s", amiName)
	aws.MakeTestPacker(0, data, confDir)

	rc := realMain([]string{"--quiet", "--config-dir", confDir, "baseami", "--region", region, releasefile, version})
	if rc != 0 {
		return errors.New("Build image failed")
	}
	return nil
}

func loadAmiFile(filepath string) (map[string]string, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func TestBuildImageTop(t *testing.T) {
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", awsKeyID)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	awsSecretKeyID := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecretKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somevalue")

	confDir, _ := ioutil.TempDir("", "stardogtests")
	defer os.RemoveAll(confDir)

	err := buildImage("ami-nopexxxx", confDir, "4.2", "/etc/group", "us-west-1")
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}
	err = buildImage("ami-beefeast", confDir, "4.2", "/etc/group", "us-east-1")
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}
	err = buildImage("ami-beefwest", confDir, "4.2", "/etc/group", "us-west-1")
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}
	err = buildImage("ami-43xxeast", confDir, "4.3", "/etc/group", "us-east-1")
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}

	ami42Path := path.Join(confDir, "amis-4.2.json")
	if !sdutils.PathExists(ami42Path) {
		t.Fatalf("The 4.2 file doesn't exist %s", ami42Path)
	}
	m42, err := loadAmiFile(ami42Path)
	if err != nil {
		t.Fatalf("Failed to load the 4.2 ami info %s", err)
	}
	ent, ok := m42["us-west-1"]
	if !ok {
		t.Fatal("The expected region was not there")
	}
	if ent != "ami-beefwest" {
		t.Fatal("The wrong ami was found")
	}
	ent, ok = m42["us-east-1"]
	if !ok {
		t.Fatal("The expected region was not there")
	}
	if ent != "ami-beefeast" {
		t.Fatal("The wrong ami was found")
	}

	ami43Path := path.Join(confDir, "amis-4.3.json")
	if !sdutils.PathExists(ami43Path) {
		t.Fatalf("The 4.3 file doesn't exist %s", ami43Path)
	}
	m43, err := loadAmiFile(ami43Path)
	if err != nil {
		t.Fatalf("Failed to load the 4.3 ami info %s", err)
	}
	ent, ok = m43["us-east-1"]
	if !ok {
		t.Fatal("The expected region was not there")
	}
	if ent != "ami-43xxeast" {
		t.Fatal("The wrong ami was found")
	}
}

func TestBasicDeploy(t *testing.T) {
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", awsKeyID)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	awsSecretKeyID := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecretKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somevalue")

	confDir, _ := ioutil.TempDir("", "stardogtests")
	defer os.RemoveAll(confDir)
	sshKeyFile := path.Join(confDir, "keyfile")
	fakeOut := "xxx"
	ioutil.WriteFile(sshKeyFile, []byte(fakeOut), 0600)

	region := "us-west-1"
	err := buildImage("ami-deadbeef", confDir, "50.10", "/etc/group", region)
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exeDir, _, _ := aws.MakeTestTerraform(0, goodoutput1, "")
	defer os.RemoveAll(exeDir)
	os.Setenv("PATH", strings.Join([]string{exeDir, os.Getenv("PATH")}, ":"))

	consoleLog := path.Join(confDir, "output2")

	depName := randDeployName()
	rc := realMain([]string{"--config-dir", confDir, "launch", "--no-wait",
		"--region", region, "--aws-key-name", "keyname", "--private-key", sshKeyFile,
		"--sd-version", "50.10", "--license", "/etc/group", depName})
	if rc != 0 {
		t.Fatal("launch failed")
	}
	rc = realMain([]string{"--config-dir", confDir, "status", depName})
	if rc != 0 {
		t.Fatal("status failed")
	}
	if !sdutils.PathExists(path.Join(confDir, "deployments", depName)) {
		t.Fatal("The deployment directory should exist")
	}
	rc = realMain([]string{"--console-file", consoleLog, "--config-dir", confDir, "volume", "status", depName})
	if rc != 0 {
		t.Fatal("volume status failed")
	}
	b, err := ioutil.ReadFile(consoleLog)
	if err != nil {
		t.Fatal("the console log was not created")
	}
	vols := []string{"vol-46313ce8", "vol-a34ba11c", "vol-50313cfe"}
	output1 := string(b)
	for _, vol := range vols {
		if !strings.Contains(output1, vol) {
			t.Fatalf("The volume %s should have been found", vol)
		}
	}
	rc = realMain([]string{"--console-file", consoleLog, "--config-dir", confDir, "instance", "status", depName})
	if rc != 0 {
		t.Fatal("instance status failed")
	}
	b, err = ioutil.ReadFile(consoleLog)
	if err != nil {
		t.Fatal("the console log was not created")
	}
	instances := []string{"stardog.sometest.com", "bastion.sometest.com"}
	output1 = string(b)
	for _, inst := range instances {
		if !strings.Contains(output1, inst) {
			t.Fatalf("The instance %s should have been found", inst)
		}
	}

	rc = realMain([]string{"--config-dir", confDir, "destroy", "--force", depName})
	if rc != 0 {
		t.Fatal("status failed")
	}
	if sdutils.PathExists(path.Join(confDir, "deployments", depName)) {
		t.Fatal("The deployment directory should not exist")
	}
}

func TestSteppedOutDeploy(t *testing.T) {
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	defer os.Setenv("AWS_ACCESS_KEY_ID", awsKeyID)
	os.Setenv("AWS_ACCESS_KEY_ID", "gravitontest")
	awsSecretKeyID := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecretKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", "somevalue")

	confDir, _ := ioutil.TempDir("", "stardogtests")
	defer os.RemoveAll(confDir)
	sshKeyFile := path.Join(confDir, "keyfile")
	fakeOut := "xxx"
	ioutil.WriteFile(sshKeyFile, []byte(fakeOut), 0600)

	region := "us-west-1"
	err := buildImage("ami-deadbeef", confDir, "4.2", "/etc/group", region)
	if err != nil {
		t.Fatalf("Failed to make the ami %s", err)
	}

	startPath := os.Getenv("PATH")
	defer os.Setenv("PATH", startPath)
	exeDir, _, _ := aws.MakeTestTerraform(0, goodoutput1, "")
	defer os.RemoveAll(exeDir)
	sshExeDir, _, _ := aws.MakeTestSSH(0)
	defer os.RemoveAll(sshExeDir)
	os.Setenv("PATH", strings.Join([]string{exeDir, os.Getenv("PATH")}, ":"))

	depName := randDeployName()
	rc := realMain([]string{"--config-dir", confDir, "deployment", "new", "--private-key", sshKeyFile, "--aws-key-name", "keyname", "--region", region, depName, "4.2"})
	if rc != 0 {
		t.Fatal("depl new failed")
	}
	rc = realMain([]string{"--config-dir", confDir, "volume", "new", depName, "/etc/group", "1", "1"})
	if rc != 0 {
		t.Fatal("status failed")
	}
	if !sdutils.PathExists(path.Join(confDir, "deployments", depName)) {
		t.Fatal("The deployment directory should exist")
	}
	os.Setenv("STARDOG_GRAVITON_HEALTHY", "false")
	rc = realMain([]string{"--config-dir", confDir, "instance", "new", "--wait-timeout", "2", depName, "1"})
	os.Unsetenv("STARDOG_GRAVITON_HEALTHY")
	if rc == 0 {
		t.Fatal("instance start should have failed with timeout")
	}
	rc = realMain([]string{"--config-dir", confDir, "instance", "status", depName})
	if rc != 0 {
		t.Fatal("instance status failed")
	}
	rc = realMain([]string{"--config-dir", confDir, "instance", "destroy", "--force", depName})
	if rc != 0 {
		t.Fatal("Should be able to destroy the instance")
	}
	rc = realMain([]string{"--config-dir", confDir, "volume", "destroy", "--force", depName})
	if rc != 0 {
		t.Fatal("Should be able to destroy the volume")
	}
	rc = realMain([]string{"--config-dir", confDir, "deployment", "destroy", "--force", depName})
	if rc != 0 {
		t.Fatal("Should be able to destroy the deployment")
	}

	if sdutils.PathExists(path.Join(confDir, "deployments", depName)) {
		t.Fatal("The deployment directory should not exist")
	}
}

func TestDestroyNoExist(t *testing.T) {
	confDir, _ := ioutil.TempDir("", "stardogtests")
	defer os.RemoveAll(confDir)

	depName := randDeployName()
	rc := realMain([]string{"--config-dir", confDir, "deployment", "destroy", "--force", depName})
	if rc == 0 {
		t.Fatal("Should fail when deleting a non existing deployment")
	}
	rc = realMain([]string{"--config-dir", confDir, "volume", "destroy", "--force", depName})
	if rc == 0 {
		t.Fatal("Should fail when deleting a non existing volume")
	}
	rc = realMain([]string{"--config-dir", confDir, "instance", "destroy", "--force", depName})
	if rc == 0 {
		t.Fatal("Should fail when deleting a non existing instance")
	}
	rc = realMain([]string{"--config-dir", confDir, "instance", "status", depName})
	if rc == 0 {
		t.Fatal("Should fail when status a non existing instance")
	}
	rc = realMain([]string{"--config-dir", confDir, "volume", "status", depName})
	if rc == 0 {
		t.Fatal("Should fail when status a non existing instance")
	}

	rc = realMain([]string{"--config-dir", confDir, "volume", "new", depName, "/etc/group", "/etc/group", "1", "1"})
	if rc == 0 {
		t.Fatal("Should fail when deleting a non existing volume")
	}
	rc = realMain([]string{"--config-dir", confDir, "instance", "new", depName, "1"})
	if rc == 0 {
		t.Fatal("Should fail when deleting a non existing instance")
	}
	rc = realMain([]string{"--config-dir", confDir, "client", depName, "hostname"})
	if rc == 0 {
		t.Fatal("client command should have failed")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	os.Setenv("STARDOG_GRAVITON_UNIT_TEST", "on")
	dir, err := ioutil.TempDir("", "stardogtest")
	if err != nil {
		panic("failed to make a temp dir")
	}

	_, _, err = aws.MakeTestPacker(0, "amazon-ebs,artifact,0,string,AMIs were created:ami-deadbeef", dir)
	if err != nil {
		panic("failed to make fake packer")
	}
	_, _, err = aws.MakeTestTerraform(0, "terraform data", dir)
	if err != nil {
		panic("failed to make fake terraform")
	}
	// We need to trick the tests into thinking packer and terraform exists for stubs
	os.Setenv("PATH", strings.Join([]string{dir, os.Getenv("PATH")}, ":"))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randDeployName() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestCleanup(t *testing.T) {
	os.Remove(packerFile)
	os.Remove(terraformFile)
}
