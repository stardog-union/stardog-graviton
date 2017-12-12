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
	"os"
	"os/exec"
	"path"
	"strings"

	"runtime"

	"github.com/stardog-union/stardog-graviton/sdutils"
)

var (
	baseUbuntu1604 = map[string]string{
		"eu-west-2":      "ami-03998867",
		"ap-northeast-1": "ami-0417e362",
		"ap-northeast-2": "ami-536ab33d",
		"ap-southeast-1": "ami-9f28b3fc",
		"ap-southeast-2": "ami-bb1901d8",
		"sa-east-1":      "ami-a41869c8",
		"us-east-2":      "ami-dbbd9dbe",
		"eu-west-1":      "ami-674cbc1e",
		"eu-central-1":   "ami-958128fa",
		"us-east-1":      "ami-1d4e7a66",
		"us-west-1":      "ami-969ab1f6",
		"us-west-2":      "ami-0a00ce72",
	}
	imageVersion = "3.0.0"
)

func getBaseAMI(cliContext sdutils.AppContext, region string) (string, error) {
	var err error

	baseFileName := path.Join(cliContext.GetConfigDir(), fmt.Sprintf("base-amis-%s.json", cliContext.GetVersion()))
	if !sdutils.PathExists(baseFileName) {
		err = sdutils.WriteJSON(baseUbuntu1604, baseFileName)
		if err != nil {
			return "", err
		}
	}

	baseAMIMap := make(map[string]string)
	err = sdutils.LoadJSON(&baseAMIMap, baseFileName)
	if err != nil {
		return "", nil
	}
	return baseAMIMap[region], nil
}

func lineScanner(cliContext sdutils.AppContext, line string) *sdutils.ScanResult {
	if strings.Contains(line, "amazon-ebs,artifact,0,string,AMIs were created:") {
		startNdx := strings.Index(line, "ami-")
		if startNdx == -1 {
			cliContext.Logf(sdutils.ERROR, "Did not found the AMI in the expected packer output.")
			return nil
		}
		if len(line) < startNdx+12 {
			cliContext.Logf(sdutils.ERROR, "The ami string is too short.")
		} else {
			amiID := line[startNdx : startNdx+12]
			return &sdutils.ScanResult{Key: "AMI", Value: amiID}
		}
	}
	return nil
}

func (a *awsPlugin) HaveImage(c sdutils.AppContext) bool {
	amiMap, err := loadAmiAmp(c)
	if err != nil {
		return false
	}
	_, ok := amiMap[a.Region]
	return ok
}

func (a *awsPlugin) BuildImage(context sdutils.AppContext, sdReleaseFilePath string, version string) error {
	context.Logf(sdutils.DEBUG, "Build AMI image\n")

	neededEnvs := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}
	for _, e := range neededEnvs {
		if os.Getenv(e) == "" {
			return fmt.Errorf("The environment variable %s must be set", e)
		}
	}
	packerURL := fmt.Sprintf("https://releases.hashicorp.com/packer/%s/packer_%s_%s_%s.zip", PackerVersion, PackerVersion, runtime.GOOS, runtime.GOARCH)
	err := sdutils.FindProgramVersion(context, "packer", PackerVersion, packerURL)
	if err != nil {
		return fmt.Errorf("We could not get a proper version of packer %s", err.Error())
	}

	context.Logf(sdutils.DEBUG, "Place assets\n")

	dir, err := PlaceAsset(context, context.GetConfigDir(), "etc/packer", true)
	if err != nil {
		return err
	}
	context.Logf(sdutils.DEBUG, "Extracting packer files to: %s\n", dir)
	context.ConsoleLog(2, "Extracting packer files to: %s\n", dir)
	defer os.RemoveAll(dir)

	packerPath, err := GetPackerPath(context)
	if err != nil {
		return err
	}

	ami := a.AmiID
	if ami == "" {
		ami, err = getBaseAMI(context, a.Region)
		if err != nil {
			context.Logf(sdutils.ERROR, "Failed to properly obtain the base AMI.  Failing back to hard coded version. %s\n", err.Error())
			ami = baseUbuntu1604[a.Region]
		}
	}

	workingDir := path.Join(dir, "etc/packer")
	// packer build -machine-readable -var-file vars.json stardog.json
	cmdArray := []string{packerPath, "build", "-machine-readable",
		"-var", fmt.Sprintf("stardog_release_file=%s", sdReleaseFilePath),
		"-var", fmt.Sprintf("source_ami=%s", ami),
		"-var", fmt.Sprintf("version=%s", version),
		"-var", fmt.Sprintf("region=%s", a.Region),
		"-var", fmt.Sprintf("image_version=%s", imageVersion),
		"stardog.json"}

	cmd := exec.Cmd{
		Path: cmdArray[0],
		Args: cmdArray,
		Dir:  workingDir,
	}

	context.Logf(sdutils.DEBUG, "Start packer")
	spin := sdutils.NewSpinner(context, 1, "Running packer to build the image")
	results, err := sdutils.RunCommand(context, cmd, lineScanner, spin)
	if err != nil {
		context.ConsoleLog(0, "We failed to build the image.  Please verify that you have sufficent EC2 access.")
		return err
	}
	if len(*results) < 1 {
		return fmt.Errorf("Failed to find the AMI in the packer output")
	}
	if len(*results) > 1 {
		context.Logf(sdutils.WARN, "We found more than 1 AMI")
	}

	context.ConsoleLog(1, "done\n")
	context.ConsoleLog(0, "AMI Successfully built: %s\n", (*results)[0].Value)
	context.Logf(sdutils.DEBUG, "AMI Successfully built: %s\n", (*results)[0].Value)

	amiMap, err := loadAmiAmp(context)
	if err != nil {
		return err
	}
	amiMap[a.Region] = (*results)[0].Value
	err = saveAmiMap(context, amiMap)
	if err != nil {
		return err
	}
	return nil
}
