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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"time"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

type awsDeploymentDescription struct {
	Region          string `json:"region,omitempty"`
	AmiID           string `json:"ami_id,omitempty"`
	AwsKeyName      string `json:"keyname,omitempty"`
	ZkInstanceType  string `json:"zk_instance,omitempty"`
	SdInstanceType  string `json:"sd_instance,omitempty"`
	PrivateKeyPath  string `json:"private_key_path,omitempty"`
	CreatedKey      bool   `json:"created_key,omitempty"`
	HTTPMask        string `json:"http_mask,omitempty"`
	VolumeType      string `json:"volume_type,omitempty"`
	IoPsRatio       int    `json:"iops,omitempty"`
	CustomScript    string `json:"custom_script,omitempty"`
	Version         string `json:"-"`
	Name            string `json:"-"`
	deployDir       string
	customPropFile  string
	customLog4J     string
	environment     []string
	disableSecurity bool
	ctx             sdutils.AppContext
	plugin          *awsPlugin
}

var (
	TerraformVersion = "0.8.8"
	PackerVersion    = "1.0.3"
)

func newAwsDeploymentDescription(c sdutils.AppContext, baseD *sdutils.BaseDeployment, a *awsPlugin) (*awsDeploymentDescription, error) {
	var err error
	createdKey := false

	if a.AmiID == "" {
		// If the ami is not specified look it up
		amiMap, err := loadAmiAmp(c)
		if err != nil {
			return nil, fmt.Errorf("Could not load the ami map: %s", err)
		}
		ami, ok := amiMap[a.Region]
		if !ok {
			// No ami for the deployment
			c.ConsoleLog(1, "A base AMI is required for launching the virtual appliance.  If you do not know this value you can build a new one with the 'baseami' command.\n")
			ami, err = sdutils.AskUser("Stardog base AMI", "")
			if err != nil {
				return nil, err
			}
			if ami == "" {
				return nil, fmt.Errorf("An AMI is required.  Please see the 'baseami' subcommand")
			}
		}
		a.AmiID = ami
	}
	if a.Region == "" {
		a.AwsKeyName, err = sdutils.AskUser("Region", "us-west-1")
		if err != nil {
			return nil, err
		}
	}
	if a.AwsKeyName == "" && baseD.PrivateKey == "" {
		if !c.GetInteractive() || sdutils.AskUserYesOrNo("Would you like to create an SSH key pair?") {
			newKeyName := baseD.Name + "key"
			privateKeyFilename, public, err := sdutils.GenerateKey(baseD.Directory, newKeyName)
			if err != nil {
				return nil, err
			}
			err = ImportKeyName(c, a, newKeyName, public)
			if err != nil {
				return nil, err
			}
			createdKey = true
			a.AwsKeyName = newKeyName
			baseD.PrivateKey = privateKeyFilename
		}
	}
	if a.AwsKeyName == "" {
		a.AwsKeyName, err = sdutils.AskUser("EC2 keyname", "default")
		if err != nil {
			return nil, err
		}
	}
	if baseD.PrivateKey == "" {
		baseD.PrivateKey, err = sdutils.AskUser("Private key path", "")
		if err != nil {
			return nil, err
		}
		if baseD.PrivateKey == "" {
			return nil, fmt.Errorf("A path to a private key must be provided")
		}
	}
	// Check the key name and key path
	fi, err := os.Stat(baseD.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("There was an error accessing the private key %s: %s", baseD.PrivateKey, err)
	}
	if fi.Mode()&0077 != 0 {
		return nil, fmt.Errorf("The permissions on the private key %s must only allow for user access", baseD.PrivateKey)
	}
	b, err := CheckKeyName(c, a, a.AwsKeyName)
	if err != nil {
		return nil, fmt.Errorf("There was an error checking the AWS environment: %s", err)
	}
	if !b {
		return nil, fmt.Errorf("The AWS keyname %s does not exist", a.AwsKeyName)
	}

	deployDir := sdutils.DeploymentDir(c.GetConfigDir(), baseD.Name)
	assertDir, err := PlaceAsset(c, deployDir, "etc/terraform", false)
	if err != nil {
		return nil, err
	}
	c.ConsoleLog(2, "Terraform configuration extracted to %s\n", assertDir)

	iops := a.IoPs
	if iops == 0 {
		var ok bool
		iops, ok = ValidVolumeTypes[a.VolumeType]
		if !ok {
			return nil, fmt.Errorf("%s is not a valid volume type", a.VolumeType)
		}
	}
	dd := awsDeploymentDescription{
		Region:          a.Region,
		AmiID:           a.AmiID,
		AwsKeyName:      a.AwsKeyName,
		ZkInstanceType:  a.ZkInstanceType,
		SdInstanceType:  a.SdInstanceType,
		Version:         baseD.Version,
		Name:            baseD.Name,
		PrivateKeyPath:  baseD.PrivateKey,
		ctx:             c,
		deployDir:       deployDir,
		customPropFile:  baseD.CustomPropsFile,
		customLog4J:     baseD.CustomLog4J,
		environment:     baseD.Environment,
		disableSecurity: baseD.DisableSecurity,
		CreatedKey:      createdKey,
		VolumeType:      a.VolumeType,
		IoPsRatio:       iops,
	}
	return &dd, nil
}

func (dd *awsDeploymentDescription) DestroyDeployment() error {
	if dd.CreatedKey {
		err := DeleteKeyPair(dd.ctx, dd.plugin, dd.AwsKeyName)
		return err
	}
	return nil
}

func (dd *awsDeploymentDescription) CreateVolumeSet(licensePath string, sizeOfEachVolume int, clusterSize int) error {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	return vm.CreateSet(licensePath, sizeOfEachVolume, clusterSize)
}

func (dd *awsDeploymentDescription) DeleteVolumeSet() error {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	if !vm.VolumeExists() {
		return fmt.Errorf("No volume information exists for %s", dd.Name)
	}
	return vm.DeleteSet()
}

func (dd *awsDeploymentDescription) ClusterSize() (int, error) {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	if !vm.VolumeExists() {
		return -1, fmt.Errorf("No volume information exists for %s", dd.Name)
	}
	vols, err := LoadEbsVolume(dd.ctx, vm.VolumeDir)
	if err != nil {
		return -1, err
	}
	var size int
	c, err := fmt.Sscanf(vols.ClusterSize, "%d", &size)
	if err != nil {
		return -1, err
	}
	if c != 1 {
		return -1, fmt.Errorf("Internal error: the cluster size is not coherent")
	}
	return size, nil
}

func (dd *awsDeploymentDescription) StatusVolumeSet() error {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	if !vm.VolumeExists() {
		return fmt.Errorf("No volume information exists for %s", dd.Name)
	}
	return vm.Status()
}

func (dd *awsDeploymentDescription) VolumeExists() bool {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	return vm.VolumeExists()
}

func (dd *awsDeploymentDescription) CreateInstance(volumeSize int, zookeeperSize int, idleTimeout int) error {
	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return err
	}
	return im.CreateInstance(volumeSize, zookeeperSize, idleTimeout)
}

func (dd *awsDeploymentDescription) OpenInstance(volumeSize int, zookeeperSize int, mask string, idleTimeout int) error {
	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return err
	}
	return im.OpenInstance(volumeSize, zookeeperSize, mask, idleTimeout)
}

func (dd *awsDeploymentDescription) DeleteInstance() error {
	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return err
	}
	return im.DeleteInstance()
}

func (dd *awsDeploymentDescription) StatusInstance() error {
	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return err
	}
	return im.Status()
}

func (dd *awsDeploymentDescription) FullStatus() (*sdutils.StardogDescription, error) {
	vm := NewAwsEbsVolumeManager(dd.ctx, dd)
	volumeStatus, err := vm.getStatusInformation()
	if err != nil {
		dd.ctx.ConsoleLog(1, "No volume information found %s\n", err)
	}

	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return nil, err
	}
	instS, err := getInstanceValues(im)
	if err != nil {
		dd.ctx.ConsoleLog(1, "No instance information found.\n")
	}

	sD := sdutils.StardogDescription{
		SSHHost:             im.BastionContact,
		StardogURL:          fmt.Sprintf("http://%s:5821", im.StardogContact),
		StardogInternalURL:  fmt.Sprintf("http://%s:5821", im.StardogInternalContact),
		VolumeDescription:   volumeStatus,
		InstanceDescription: instS,
		TimeStamp:           time.Now(),
	}

	return &sD, nil
}

func (dd *awsDeploymentDescription) InstanceExists() bool {
	im, err := NewEc2Instance(dd.ctx, dd)
	if err != nil {
		return false
	}
	return im.InstanceExists()
}

type awsPlugin struct {
	Region         string `json:"region,omitempty"`
	VolumeType     string `json:"volume_type,omitempty"`
	IoPs           int    `json:"iops,omitempty"`
	AmiID          string `json:"ami_id,omitempty"`
	AwsKeyName     string `json:"aws_key_name,omitempty"`
	ZkInstanceType string `json:"zk_instance_type,omitempty"`
	SdInstanceType string `json:"sd_instance_type,omitempty"`
	BastionType    string `json:"bastion_instance_type,omitempty"`
}

// GetPlugin returns the plugin interface that this module represents.
func GetPlugin() sdutils.Plugin {
	return &awsPlugin{
		Region:         "us-west-1",
		AmiID:          "",
		AwsKeyName:     "",
		ZkInstanceType: "t2.small",
		SdInstanceType: "t2.medium",
		BastionType:    "t2.small",
		VolumeType:     "gp2",
		IoPs:           0,
	}
}

func (a *awsPlugin) LoadDefaults(defaultCliOpts interface{}) error {
	// parse out from the interface any config file defaults
	b, err := json.Marshal(defaultCliOpts)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, a)
	if err != nil {
		return err
	}
	return nil
}

func (a *awsPlugin) Register(cmdOpts *sdutils.CommandOpts) error {
	cmdOpts.BuildCmd.Flag("region", fmt.Sprintf("The aws region to use [%s].", strings.Join(ValidRegions, " | "))).Default(a.Region).StringVar(&a.Region)

	cmdOpts.NewVolumesCmd.Flag("volume-type", fmt.Sprintf("The EBS volume type to use [%s]", strings.Join(GetValidVolumeTypes(), " | "))).Default(a.VolumeType).StringVar(&a.VolumeType)
	cmdOpts.NewVolumesCmd.Flag("iops", "The IOPS for the volume type").IntVar(&a.IoPs)

	cmdOpts.LaunchCmd.Flag("region", fmt.Sprintf("The aws region to use [%s]", strings.Join(ValidRegions, " | "))).Default(a.Region).StringVar(&a.Region)
	cmdOpts.LaunchCmd.Flag("zk-instance-type", "The instance type to use for zookeeper VMs").Default(a.ZkInstanceType).StringVar(&a.ZkInstanceType)
	cmdOpts.LaunchCmd.Flag("sd-instance-type", "The instance type to use for stardog VMs").Default(a.SdInstanceType).StringVar(&a.SdInstanceType)
	cmdOpts.LaunchCmd.Flag("aws-key-name", "The AWS ssh key name.").Default(a.AwsKeyName).StringVar(&a.AwsKeyName)
	cmdOpts.LaunchCmd.Flag("volume-type", fmt.Sprintf("The EBS volume type to use [%s]", strings.Join(GetValidVolumeTypes(), " | "))).Default(a.VolumeType).StringVar(&a.VolumeType)
	cmdOpts.LaunchCmd.Flag("iops", "The IOPS for the volume type").IntVar(&a.IoPs)

	cmdOpts.LeaksCmd.Flag("region", fmt.Sprintf("The aws region to use [%s]", strings.Join(ValidRegions, " | "))).Default(a.Region).StringVar(&a.Region)

	cmdOpts.NewDeploymentCmd.Flag("region", fmt.Sprintf("The aws region to use [%s].", strings.Join(ValidRegions, " | "))).Default(a.Region).StringVar(&a.Region)
	cmdOpts.NewDeploymentCmd.Flag("zk-instance-type", "The instance type to use for zookeeper VMs.").Default("m3.large").StringVar(&a.ZkInstanceType)
	cmdOpts.NewDeploymentCmd.Flag("sd-instance-type", "The instance type to use for stardog VMs.").Default("m3.large").StringVar(&a.SdInstanceType)
	cmdOpts.NewDeploymentCmd.Flag("aws-key-name", "The AWS ssh key name.").Default(a.AwsKeyName).StringVar(&a.AwsKeyName)

	return nil
}

func (a *awsPlugin) DeploymentLoader(context sdutils.AppContext, baseD *sdutils.BaseDeployment, new bool) (sdutils.Deployment, error) {
	var err error

	neededEnvs := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}
	for _, e := range neededEnvs {
		if os.Getenv(e) == "" {
			return nil, fmt.Errorf("The environment variable %s must be set", e)
		}
	}
	terraformOutputVersion := fmt.Sprintf("Terraform v%s", TerraformVersion)
	terraformURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", TerraformVersion, TerraformVersion, runtime.GOOS, runtime.GOARCH)
	err = sdutils.FindProgramVersion(context, "terraform", terraformOutputVersion, terraformURL)
	if err != nil {
		return nil, fmt.Errorf("We could not get a proper version of terraform %s", err.Error())
	}

	if new {
		awsDD, err := newAwsDeploymentDescription(context, baseD, a)
		if err != nil {
			return nil, err
		}
		awsDD.CustomScript = baseD.CustomScript
		awsDD.environment = baseD.Environment
		awsDD.disableSecurity = baseD.DisableSecurity
		baseD.CloudOpts = awsDD
		data, err := json.Marshal(baseD)
		if err != nil {
			return nil, err
		}
		confPath := path.Join(awsDD.deployDir, "config.json")
		err = ioutil.WriteFile(confPath, data, 0600)
		if err != nil {
			return nil, err
		}
		awsDD.plugin = a
		return awsDD, nil
	}
	data, err := json.Marshal(baseD.CloudOpts)
	if err != nil {
		return nil, err
	}
	var dd awsDeploymentDescription
	err = json.Unmarshal(data, &dd)
	if err != nil {
		return nil, err
	}
	dd.Name = baseD.Name
	dd.Version = baseD.Version
	dd.ctx = context
	dd.deployDir = sdutils.DeploymentDir(context.GetConfigDir(), baseD.Name)
	dd.environment = baseD.Environment
	dd.disableSecurity = baseD.DisableSecurity
	dd.plugin = a

	return &dd, nil
}

func (a *awsPlugin) GetName() string {
	return "aws"
}

func GetGravitonDependencyExe(context sdutils.AppContext, program string) (string, error) {
	path := filepath.Join(context.GetConfigDir(), program)
	if !sdutils.PathExists(path) {
		return "", fmt.Errorf("%s is not configure correctly", program)
	}
	return path, nil
}

func GetTerraformPath(context sdutils.AppContext) (string, error) {
	return GetGravitonDependencyExe(context, "terraform")
}

func GetPackerPath(context sdutils.AppContext) (string, error) {
	return GetGravitonDependencyExe(context, "packer")
}
