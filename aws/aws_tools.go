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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/stardog-union/stardog-graviton/sdutils"
)

var (
	// ValidRegions is the list of regions that are supported by this plugin
	ValidRegions = []string{
		"us-west-1", "us-west-2", "us-east-1",
		"us-east-2", "eu-central-1", "eu-west-1",
	}
	// ValidVolumeTypes is the list of volume types that are supported by this plugin and
	// their default iops values
	ValidVolumeTypes = make(map[string]int)
)

func init() {
	ValidVolumeTypes["standard"] = 0
	ValidVolumeTypes["gp2"] = 0
	ValidVolumeTypes["io1"] = 20
}

// GetValidVolumeTypes returns a list of the volume types that are supported
func GetValidVolumeTypes() []string {
	keys := []string{}
	for k := range ValidVolumeTypes {
		keys = append(keys, k)
	}
	return keys
}

func hasTag(tags []*ec2.Tag, tagVal string, possibleDelpoyNames *map[string]bool) bool {
	for _, t := range tags {
		if t.Key != nil && *t.Key == "StardogVirtualAppliance" {
			if tagVal == "" || (t.Value != nil && tagVal == *t.Value) {
				(*possibleDelpoyNames)[*t.Value] = true
				return true
			}
		}
	}
	return false
}

func getAsgLc(c sdutils.AppContext, sess *session.Session, conf *aws.Config, tagVal string, possibleDelpoyNames *map[string]bool) ([]*autoscaling.LaunchConfiguration, []*autoscaling.Group) {
	lcList := []*autoscaling.LaunchConfiguration{}
	asgList := []*autoscaling.Group{}

	autoscaleSvc := autoscaling.New(sess, conf)
	asResp, err := autoscaleSvc.DescribeAutoScalingGroups(nil)
	if err != nil {
		c.ConsoleLog(1, "Failed to get autoscale groups: %s\n", err)
		return lcList, asgList
	}
	for _, g := range asResp.AutoScalingGroups {
		for _, t := range g.Tags {
			c.Logf(sdutils.DEBUG, "Checking the ASG %s\n", *g.AutoScalingGroupName)
			if t.Key != nil && *t.Key == "StardogVirtualAppliance" {
				if tagVal == "" || (t.Value != nil && tagVal == *t.Value) {
					c.Logf(sdutils.DEBUG, "Found ASG %s\n", *g.AutoScalingGroupName)
					(*possibleDelpoyNames)[*t.Value] = true
					asgList = append(asgList, g)
					if g.LaunchConfigurationName != nil {
						lcInput := autoscaling.DescribeLaunchConfigurationsInput{
							LaunchConfigurationNames: []*string{g.LaunchConfigurationName}}
						lcResp, err := autoscaleSvc.DescribeLaunchConfigurations(&lcInput)
						if err != nil {
							c.Logf(sdutils.WARN, "Failed to describe the launch configuration %s: %s", *g.LaunchConfigurationName, err)
						} else {
							lcList = append(lcList, lcResp.LaunchConfigurations...)
						}
					}
				}
			}
		}
	}
	return lcList, asgList
}

func destroyAsgLc(c sdutils.AppContext, sess *session.Session, conf *aws.Config, lcList []*autoscaling.LaunchConfiguration, asgList []*autoscaling.Group) error {
	c.ConsoleLog(1, "Destroying the autoscaling groups\n")
	autoscaleSvc := autoscaling.New(sess, conf)
	for _, asg := range asgList {
		c.ConsoleLog(2, "Destroying %s\n", *asg.AutoScalingGroupName)
		input := autoscaling.DeleteAutoScalingGroupInput{AutoScalingGroupName: asg.AutoScalingGroupName}
		_, err := autoscaleSvc.DeleteAutoScalingGroup(&input)
		if err != nil {
			c.Logf(sdutils.WARN, "Failed to delete the ASG %s, %s", *asg.AutoScalingGroupName, err)
			c.ConsoleLog(1, "Failed to delete the ASG %s, %s\n", *asg.AutoScalingGroupName, err)
		}
	}
	c.ConsoleLog(1, "Destroying the launch configurations\n")
	for _, lc := range lcList {
		c.ConsoleLog(2, "Destroying %s\n", *lc.LaunchConfigurationName)
		input := autoscaling.DeleteLaunchConfigurationInput{LaunchConfigurationName: lc.LaunchConfigurationName}
		_, err := autoscaleSvc.DeleteLaunchConfiguration(&input)
		if err != nil {
			c.Logf(sdutils.WARN, "Failed to delete the LC %s, %s", *lc.LaunchConfigurationName, err)
			c.ConsoleLog(1, "Failed to delete the LC %s, %s\n", *lc.LaunchConfigurationName, err)
		}
	}
	return nil
}

func getInstances(c sdutils.AppContext, sess *session.Session, conf *aws.Config, tagVal string, possibleDelpoyNames *map[string]bool) []*ec2.Instance {
	instList := []*ec2.Instance{}
	svc := ec2.New(sess, conf)
	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		c.ConsoleLog(1, "Failed to get any instances: %s\n", err)
		return instList
	}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if hasTag(inst.Tags, tagVal, possibleDelpoyNames) && *inst.State.Code != 48 {
				instList = append(instList, inst)
				c.Logf(sdutils.DEBUG, "Found instance %s with tag %s", *inst.InstanceId, tagVal)
			}
		}
	}
	return instList
}

func getAmiVersion(c sdutils.AppContext, sess *session.Session, conf *aws.Config, ami *string) (*string, error) {
	input := ec2.DescribeImagesInput{ImageIds: []*string{ami}}
	svc := ec2.New(sess, conf)
	output, err := svc.DescribeImages(&input)
	if err != nil {
		return nil, err
	}
	if len(output.Images) != 1 {
		return nil, fmt.Errorf("No images were found with that name")
	}
	tags := output.Images[0].Tags
	if tags == nil || len(tags) < 1 {
		return nil, fmt.Errorf("No version tag found for AMI %s", *ami)
	}
	// find version
	for _, t := range tags {
		if t != nil && *t.Key == "ImageVersion" {
			return t.Value, nil
		}
	}
	return nil, nil
}

// CheckKeyName will return true or false based on the existance of the keyname in the
// configured AWS environment.  If an error occurs while communicating with AWS an
// error will be returned.
func CheckKeyName(c sdutils.AppContext, a *awsPlugin, keyname string) (bool, error) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "gravitontest" {
		return true, nil
	}
	conf := aws.Config{Region: aws.String(a.Region)}
	sess, err := session.NewSession()
	if err != nil {
		return false, err
	}

	svc := ec2.New(sess, &conf)
	keyOut, err := svc.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return false, err
	}

	for _, s := range keyOut.KeyPairs {
		if s.KeyName != nil && *s.KeyName == keyname {
			return true, nil
		}
	}
	return false, nil
}

func ImportKeyName(c sdutils.AppContext, a *awsPlugin, keyname string, publickey []byte) error {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "gravitontest" {
		return nil
	}
	conf := aws.Config{Region: aws.String(a.Region)}
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	svc := ec2.New(sess, &conf)

	keyInput := ec2.ImportKeyPairInput{
		KeyName:           &keyname,
		PublicKeyMaterial: publickey,
	}
	_, err = svc.ImportKeyPair(&keyInput)
	if err != nil {
		return err
	}
	return nil
}

func DeleteKeyPair(c sdutils.AppContext, a *awsPlugin, keyname string) error {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "gravitontest" {
		return nil
	}
	conf := aws.Config{Region: aws.String(a.Region)}
	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	svc := ec2.New(sess, &conf)
	_, err = svc.DeleteKeyPair(&ec2.DeleteKeyPairInput{KeyName: &keyname})
	if err != nil {
		return err
	}
	return nil
}

func destroyInstances(c sdutils.AppContext, sess *session.Session, conf *aws.Config, instList []*ec2.Instance) error {
	svc := ec2.New(sess, conf)
	for _, inst := range instList {
		input := ec2.TerminateInstancesInput{InstanceIds: []*string{inst.InstanceId}}
		svc.TerminateInstances(&input)
	}
	return nil
}

func getSecurityGroups(c sdutils.AppContext, sess *session.Session, conf *aws.Config, tagVal string, possibleDelpoyNames *map[string]bool) []*ec2.SecurityGroup {
	sgList := []*ec2.SecurityGroup{}
	svc := ec2.New(sess, conf)
	sgoa, err := svc.DescribeSecurityGroups(nil)
	if err != nil {
		c.ConsoleLog(1, "Failed to get any security groups: %s\n", err)
		return sgList
	}

	for _, sg := range sgoa.SecurityGroups {
		if hasTag(sg.Tags, tagVal, possibleDelpoyNames) {
			c.Logf(sdutils.DEBUG, "Found security group %s", *sg.GroupName)
			sgList = append(sgList, sg)
		}
	}
	return sgList
}

func destroySecurityGroups(c sdutils.AppContext, sess *session.Session, conf *aws.Config, sgList []*ec2.SecurityGroup) error {
	svc := ec2.New(sess, conf)
	for _, sg := range sgList {
		input := ec2.DeleteSecurityGroupInput{GroupId: sg.GroupId}
		_, err := svc.DeleteSecurityGroup(&input)
		if err != nil {
			c.Logf(sdutils.WARN, "Failed to delete the security group %s.  %s", *sg.GroupName, err)
			c.ConsoleLog(1, "Failed to delete the security group %s. %s\n", *sg.GroupName, err)
		}
	}
	return nil
}

func getElbs(c sdutils.AppContext, sess *session.Session, conf *aws.Config, tagVal string) []*elb.LoadBalancerDescription {
	elbList := []*elb.LoadBalancerDescription{}
	svc := elb.New(sess, conf)
	resp, err := svc.DescribeLoadBalancers(nil)
	if err != nil {
		c.ConsoleLog(1, "Failed to get any load balancers: %s\n", err)
		return elbList
	}
	for _, elbD := range resp.LoadBalancerDescriptions {
		if elbD.LoadBalancerName != nil && strings.Contains(*elbD.LoadBalancerName, tagVal) {
			elbList = append(elbList, elbD)
		}
	}
	return elbList
}

func destroyLoadBalancers(c sdutils.AppContext, sess *session.Session, conf *aws.Config, elbList []*elb.LoadBalancerDescription) error {
	svc := elb.New(sess, conf)
	for _, e := range elbList {
		input := elb.DeleteLoadBalancerInput{LoadBalancerName: e.LoadBalancerName}
		_, err := svc.DeleteLoadBalancer(&input)
		if err != nil {
			c.Logf(sdutils.WARN, "Failed to delete the load balancer %s", *e.LoadBalancerName)
			c.ConsoleLog(1, "Failed to delete the load balancer %s\n", *e.LoadBalancerName)
		}
	}
	return nil
}

func (a *awsPlugin) FindLeaks(c sdutils.AppContext, deploymentName string, destroy bool, force bool) error {
	possibleDeployNames := make(map[string]bool)

	if deploymentName != "" {
		possibleDeployNames[deploymentName] = true
	}

	conf := aws.Config{Region: aws.String(a.Region)}
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	c.ConsoleLog(1, "Looking for AWS resources\n")

	lcList, asgList := getAsgLc(c, sess, &conf, deploymentName, &possibleDeployNames)
	instList := getInstances(c, sess, &conf, deploymentName, &possibleDeployNames)
	sgList := getSecurityGroups(c, sess, &conf, deploymentName, &possibleDeployNames)

	elbList := []*elb.LoadBalancerDescription{}
	for tagName := range possibleDeployNames {
		tmpElbList := getElbs(c, sess, &conf, tagName)
		elbList = append(elbList, tmpElbList...)
	}

	c.ConsoleLog(1, "Found %d autoscaling groups\n", len(asgList))
	for _, asg := range asgList {
		c.ConsoleLog(1, "\t%s\n", *asg.AutoScalingGroupName)
	}
	c.ConsoleLog(1, "Found %d launch configurations\n", len(lcList))
	for _, lc := range lcList {
		c.ConsoleLog(1, "\t%s\n", *lc.LaunchConfigurationName)
	}
	c.ConsoleLog(1, "Found %d load balancers\n", len(elbList))
	for _, elb := range elbList {
		c.ConsoleLog(1, "\t%s\n", *elb.LoadBalancerName)
	}
	c.ConsoleLog(1, "Found %d instances\n", len(instList))
	for _, inst := range instList {
		c.ConsoleLog(1, "\t%s\n", *inst.InstanceId)
	}
	c.ConsoleLog(1, "Found %d security groups\n", len(sgList))
	for _, sg := range sgList {
		c.ConsoleLog(1, "\t%s\n", *sg.GroupName)
	}

	if !destroy {
		return nil
	}
	if !force {
		if !sdutils.AskUserYesOrNo("Would you like to destroy these resources?") {
			return nil
		}
	}
	destroyInstances(c, sess, &conf, instList)
	destroyAsgLc(c, sess, &conf, lcList, asgList)
	destroyLoadBalancers(c, sess, &conf, elbList)
	destroySecurityGroups(c, sess, &conf, sgList)

	return nil
}

func amiFileName(cliContext sdutils.AppContext) string {
	return path.Join(cliContext.GetConfigDir(), fmt.Sprintf("amis-%s.json", cliContext.GetVersion()))
}

func loadAmiAmp(cliContext sdutils.AppContext) (map[string]string, error) {
	amiMap := make(map[string]string)
	amiMapFile := amiFileName(cliContext)
	cliContext.Logf(sdutils.DEBUG, "Loading the AMI file %s\n", amiMapFile)
	if _, err := os.Stat(amiMapFile); err == nil {
		cliContext.Logf(sdutils.DEBUG, "Read AMI file\n")
		data, err := ioutil.ReadFile(amiMapFile)
		if err != nil {
			return nil, err
		}
		cliContext.Logf(sdutils.DEBUG, "Unmarshall\n")

		err = json.Unmarshal(data, &amiMap)
		if err != nil {
			return nil, err
		}
	}
	cliContext.Logf(sdutils.DEBUG, "Got the ami map %s\n", amiMap)
	return amiMap, nil
}

func saveAmiMap(cliContext sdutils.AppContext, amiMap map[string]string) error {
	data, err := json.Marshal(&amiMap)
	if err != nil {
		return err
	}
	amiMapFile := amiFileName(cliContext)
	cliContext.Logf(sdutils.DEBUG, "Saving the AMI file %s\n", amiMapFile)
	err = ioutil.WriteFile(amiMapFile, data, 0600)
	if err != nil {
		return err
	}
	return err
}

// PlaceAsset will write data that was compiled in with go-bindata to a file.
func PlaceAsset(cliContext sdutils.AppContext, dir string, assentName string, temp bool) (string, error) {
	var err error
	if temp {
		dir, err = ioutil.TempDir(dir, "stardog")
		if err != nil {
			return "", err
		}
	} else {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				cliContext.ConsoleLog(0, "ERROR %s\n", err.Error())
				return "", err
			}
		}
	}

	err = RestoreAssets(dir, assentName)
	if err != nil {
		fmt.Println("asset not found")
		return "", err
	}
	return dir, nil
}
