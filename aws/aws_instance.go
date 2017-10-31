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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

// Ec2Instance represents an instance of a Stardog service in AWS.
type Ec2Instance struct {
	DeploymentName         string             `json:"deployment_name,omitempty"`
	Region                 string             `json:"aws_region,omitempty"`
	KeyName                string             `json:"aws_key_name,omitempty"`
	Version                string             `json:"version,omitempty"`
	ZkInstanceType         string             `json:"zk_instance_type,omitempty"`
	SdInstanceType         string             `json:"stardog_instance_type,omitempty"`
	ZkSize                 string             `json:"zookeeper_size,omitempty"`
	SdSize                 string             `json:"stardog_size,omitempty"`
	AmiID                  string             `json:"baseami,omitempty"`
	PrivateKey             string             `json:"private_key,omitempty"`
	HTTPMask               string             `json:"http_subnet,omitempty"`
	ELBIdleTimeout         string             `json:"elb_idle_timeout,omitempty"`
	CustomPropsData        string             `json:"custom_properties_data,omitempty"`
	CustomLog4JData        string             `json:"custom_log4j_data,omitempty"`
	Environment            string             `json:"environment_variables,omitempty"`
	StartOpts              string             `json:"stardog_start_opts,omitempty"`
	RootVolumeSize         int                `json:"root_volume_size"`
	RootVolumeType         string             `json:"root_volume_type"`
	CustomScript           string             `json:"custom_script,omitempty"`
	DeployDir              string             `json:"-"`
	Ctx                    sdutils.AppContext `json:"-"`
	BastionContact         string             `json:"-"`
	StardogContact         string             `json:"-"`
	StardogInternalContact string             `json:"-"`
	ZkNodesContact         []string           `json:"-"`
}

// InstanceStatusDescription describes details about a running Stardog instance.
// The zookeeper contact strings are described.
type InstanceStatusDescription struct {
	ZkNodesContact []string
}

// NewEc2Instance instanciates a AwsEc2Instance object which will be used to boot or
// inspect a Stardog ec2 instance.
func NewEc2Instance(ctx sdutils.AppContext, dd *awsDeploymentDescription) (*Ec2Instance, error) {
	var err error
	customData := ""
	if dd.customPropFile != "" {
		data, err := ioutil.ReadFile(dd.customPropFile)
		if err != nil {
			return nil, fmt.Errorf("Invalid custom properties file: %s", err)
		}
		customData = string(data)
	}

	customLog4J := ""
	if dd.customLog4J != "" {
		log4JData, err := ioutil.ReadFile(dd.customLog4J)
		if err != nil {
			return nil, fmt.Errorf("Invalid custom properties file: %s", err)
		}
		customLog4J = base64.StdEncoding.EncodeToString(log4JData)
	}

	var envBuffer bytes.Buffer
	for _, env := range dd.environment {
		envBuffer.WriteString(fmt.Sprintf("export %s\n", env))
	}
	// The custom script cannot be null in terraform so make a temp one
	if dd.CustomScript == "" {
		dd.CustomScript = path.Join(dd.deployDir, "dummy.sh")
		tmpScript := "#!/bin/bash\nexit 0\n"
		err = ioutil.WriteFile(dd.CustomScript, []byte(tmpScript), 0755)
		if err != nil {
			return nil, err
		}
	}
	instance := Ec2Instance{
		DeploymentName:  dd.Name,
		Region:          dd.Region,
		KeyName:         dd.AwsKeyName,
		Version:         dd.Version,
		ZkInstanceType:  dd.ZkInstanceType,
		SdInstanceType:  dd.SdInstanceType,
		AmiID:           dd.AmiID,
		PrivateKey:      dd.PrivateKeyPath,
		DeployDir:       dd.deployDir,
		CustomScript:    dd.CustomScript,
		Ctx:             ctx,
		CustomPropsData: customData,
		CustomLog4JData: customLog4J,
		Environment:     envBuffer.String(),
	}
	if dd.disableSecurity {
		instance.StartOpts = "--disable-security"
	}
	return &instance, nil
}

func volumeLineScanner(cliContext sdutils.AppContext, line string) *sdutils.ScanResult {
	outputKeys := []string{"load_balancer_ip"}

	for _, k := range outputKeys {
		if strings.HasPrefix(line, k) {
			la := strings.Split(line, " = ")
			return &sdutils.ScanResult{Key: la[0], Value: la[1]}
		}
	}
	return nil
}

func (awsI *Ec2Instance) runTerraformApply(volumeSize int, zookeeperSize int, mask string, idleTimeout int, message string) error {
	awsI.ZkSize = fmt.Sprintf("%d", zookeeperSize)
	awsI.ELBIdleTimeout = fmt.Sprintf("%d", idleTimeout)

	vol, err := LoadEbsVolume(awsI.Ctx, path.Join(awsI.DeployDir, "etc", "terraform", "volumes"))
	if err != nil {
		return err
	}

	awsI.SdSize = vol.ClusterSize
	awsI.HTTPMask = mask
	awsI.RootVolumeType = "standard"
	awsI.RootVolumeSize = volumeSize

	instanceWorkingDir := path.Join(awsI.DeployDir, "etc", "terraform", "instance")
	instanceConfPath := path.Join(instanceWorkingDir, "instance.json")
	if sdutils.PathExists(instanceConfPath) && mask == "" {
		awsI.Ctx.ConsoleLog(1, "The instance already exists.\n")
		awsI.Ctx.Logf(sdutils.INFO, "The instance already exists.")
	}
	err = sdutils.WriteJSON(awsI, instanceConfPath)
	if err != nil {
		return err
	}

	terraformPath, err := GetTerraformPath(awsI.Ctx)
	if err != nil {
		return err
	}

	cmdArray := []string{terraformPath, "apply", "-var-file",
		instanceConfPath}
	cmd := exec.Cmd{
		Path: cmdArray[0],
		Args: cmdArray,
		Dir:  instanceWorkingDir,
	}
	awsI.Ctx.Logf(sdutils.INFO, "Running terraform...\n")
	spin := sdutils.NewSpinner(awsI.Ctx, 1, message)
	_, err = sdutils.RunCommand(awsI.Ctx, cmd, volumeLineScanner, spin)
	if err != nil {
		return err
	}
	return nil
}

// CreateInstance will boot up a Stardog service in AWS.
func (awsI *Ec2Instance) CreateInstance(volumeSize int, zookeeperSize int, idleTimeout int) error {
	err := awsI.runTerraformApply(volumeSize, zookeeperSize, "0.0.0.0/32", idleTimeout, "Creating the instance VMs...")
	if err != nil {
		awsI.Ctx.ConsoleLog(1, "Failed to create the instance.\n")
		return err
	}
	awsI.Ctx.ConsoleLog(1, "Successfully created the instance.\n")
	return nil
}

// OpenInstance will open the firewall to allow incoming traffic to port 5821 from
// the give CIDR.
func (awsI *Ec2Instance) OpenInstance(volumeSize int, zookeeperSize int, mask string, idleTimeout int) error {
	err := awsI.runTerraformApply(volumeSize, zookeeperSize, mask, idleTimeout, "Opening the firewall...")
	if err != nil {
		awsI.Ctx.ConsoleLog(1, "Failed to open up the instance.\n")
		return err
	}
	awsI.Ctx.ConsoleLog(1, "Successfully opened up the instance.\n")
	return nil
}

// DeleteInstance will teardown the Stardog service.
func (awsI *Ec2Instance) DeleteInstance() error {
	instanceWorkingDir := path.Join(awsI.DeployDir, "etc", "terraform", "instance")
	instanceConfPath := path.Join(instanceWorkingDir, "instance.json")
	if !sdutils.PathExists(instanceConfPath) {
		return fmt.Errorf("There is no configured instance")
	}
	terraformPath, err := GetTerraformPath(awsI.Ctx)
	if err != nil {
		return err
	}
	cmdArray := []string{terraformPath, "destroy", "-force", "-var-file", instanceConfPath}
	cmd := exec.Cmd{
		Path: cmdArray[0],
		Args: cmdArray,
		Dir:  instanceWorkingDir,
	}
	awsI.Ctx.Logf(sdutils.INFO, "Running terraform...\n")
	spin := sdutils.NewSpinner(awsI.Ctx, 1, "Deleting the instance VMs")
	_, err = sdutils.RunCommand(awsI.Ctx, cmd, volumeLineScanner, spin)
	if err != nil {
		return err
	}
	os.Remove(instanceConfPath)
	awsI.Ctx.ConsoleLog(1, "Successfully destroyed the instance.\n")
	return nil
}

// InstanceExists will return a bool if the associated AwsEc2Instance has already been
// created.
func (awsI *Ec2Instance) InstanceExists() bool {
	instanceWorkingDir := path.Join(awsI.DeployDir, "etc", "terraform", "instance")
	instanceConfPath := path.Join(instanceWorkingDir, "instance.json")
	return sdutils.PathExists(instanceConfPath)
}

// OutputEntry allows the plugin to return opaque information and mark it as sensitive
// or not.  If it is sensitive the base code knows not to print it out or write it to a
// log.
type OutputEntry struct {
	Sensitive bool        `json:"sensitive,omitempty"`
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

func getInstanceValues(awsI *Ec2Instance) (*InstanceStatusDescription, error) {
	instanceWorkingDir := path.Join(awsI.DeployDir, "etc", "terraform", "instance")
	instanceConfPath := path.Join(instanceWorkingDir, "instance.json")
	if !sdutils.PathExists(instanceConfPath) {
		return nil, fmt.Errorf("There is no configured instance")
	}
	terraformPath, err := GetTerraformPath(awsI.Ctx)
	if err != nil {
		return nil, err
	}
	cmdArray := []string{terraformPath, "output", "-json"}
	cmd := exec.Cmd{
		Path: cmdArray[0],
		Args: cmdArray,
		Dir:  instanceWorkingDir,
	}
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	try := make(map[string]OutputEntry)
	err = json.Unmarshal(data, &try)
	if err != nil {
		return nil, err
	}

	awsI.StardogInternalContact = try["stardog_internal_contact"].Value.(string)
	awsI.StardogContact = try["stardog_contact"].Value.(string)
	awsI.BastionContact = try["bastion_contact"].Value.(string)
	interList := try["zookeeper_nodes"].Value.([]interface{})
	awsI.ZkNodesContact = make([]string, len(interList), len(interList))
	for ndx, x := range interList {
		awsI.ZkNodesContact[ndx] = x.(string)
	}

	s := InstanceStatusDescription{
		ZkNodesContact: awsI.ZkNodesContact,
	}

	return &s, nil
}

// Status will print the status of the ec2 instance.
func (awsI *Ec2Instance) Status() error {
	_, err := getInstanceValues(awsI)
	if err != nil {
		return err
	}

	awsI.Ctx.ConsoleLog(1, "Stardog: %s\n", fmt.Sprintf("http://%s:5821", awsI.StardogContact))
	awsI.Ctx.ConsoleLog(1, "SSH: %s\n", awsI.BastionContact)
	return nil
}
