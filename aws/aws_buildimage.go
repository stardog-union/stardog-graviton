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

	"github.com/stardog-union/stardog-graviton/sdutils"
)

var (
	baseUbuntu1604 = map[string]string{
		"ap-northeast-1": "ami-31892c50",
		"ap-southeast-1": "ami-18e7417b",
		"ap-southeast-2": "ami-7be4d618",
		"cn-north-1":     "ami-d7c511ba",
		"eu-central-1":   "ami-597c8236",
		"eu-west-1":      "ami-c593deb6",
		"sa-east-1":      "ami-909b06fc",
		"us-east-1":      "ami-fd6e3bea",
		"us-east-2":      "ami-0a104a6f",
		"us-gov-west-1":  "ami-8df24aec",
		"us-west-1":      "ami-73531b13",
		"us-west-2":      "ami-f1ca1091",
	}
)

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

	context.Logf(sdutils.DEBUG, "Place assets\n")

	dir, err := PlaceAsset(context, context.GetConfigDir(), "etc/packer", true)
	if err != nil {
		return err
	}
	context.Logf(sdutils.DEBUG, "Extracting packer files to: %s\n", dir)
	context.ConsoleLog(2, "Extracting packer files to: %s\n", dir)
	defer os.RemoveAll(dir)

	packerPath, err := exec.LookPath("packer")
	if err != nil {
		return err
	}

	ami := a.AmiID
	if ami == "" {
		ami = baseUbuntu1604[a.Region]
	}

	workingDir := path.Join(dir, "etc/packer")
	// packer build -machine-readable -var-file vars.json stardog.json
	cmdArray := []string{packerPath, "build", "-machine-readable",
		"-var", fmt.Sprintf("stardog_release_file=%s", sdReleaseFilePath),
		"-var", fmt.Sprintf("source_ami=%s", ami),
		"-var", fmt.Sprintf("version=%s", version),
		"-var", fmt.Sprintf("region=%s", a.Region),
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
