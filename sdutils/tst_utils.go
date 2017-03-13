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
	"fmt"
)

type TestContext struct {
	ConfigDir string
	Version   string
}

func (c *TestContext) Logf(level int, format string, v ...interface{}) {
}

func (c *TestContext) ConsoleLog(level int, format string, v ...interface{}) {
}

func (c *TestContext) GetConfigDir() string {
	return c.ConfigDir
}

func (c *TestContext) GetVersion() string {
	return c.Version
}

func (c *TestContext) HighlightString(a ...interface{}) string {
	return fmt.Sprint(a...)
}

func (c *TestContext) SuccessString(a ...interface{}) string {
	return fmt.Sprint(a...)
}

func (c *TestContext) FailString(a ...interface{}) string {
	return fmt.Sprint(a...)
}

type TstPlugin struct {
	HasImage bool
	Dep      Deployment
}

func (tp *TstPlugin) Register(cmdOpts *CommandOpts) error {
	return nil
}

func (tp *TstPlugin) DeploymentLoader(context AppContext, baseD *BaseDeployment, new bool) (Deployment, error) {
	if tp.Dep == nil {
		return nil, fmt.Errorf("This test plugin loads no deployment")
	}
	return tp.Dep, nil
}

func (tp *TstPlugin) LoadDefaults(defaultCliOpts interface{}) error {
	return nil
}

func (tp *TstPlugin) BuildImage(context AppContext, sdReleaseFilePath string, version string) error {
	return nil
}

func (tp *TstPlugin) GetName() string {
	return "test1loaderror"
}

func (tp *TstPlugin) FindLeaks(context AppContext, deploymentName string, destroy bool, force bool) error {
	return nil
}

func (tp *TstPlugin) HaveImage(context AppContext) bool {
	return tp.HasImage
}

type TpDeployment struct {
	TstInstanceExists bool
	TstVolumeExists   bool
	SdDesc            *StardogDescription
}

func (tstDep *TpDeployment) CreateVolumeSet(licensePath string, sizeOfEachVolume int, clusterSize int) error {
	return nil
}

func (tstDep *TpDeployment) DeleteVolumeSet() error {
	return nil
}

func (tstDep *TpDeployment) StatusVolumeSet() error {
	return nil
}

func (tstDep *TpDeployment) VolumeExists() bool {
	return tstDep.TstVolumeExists
}

func (tstDep *TpDeployment) CreateInstance(zookeeperSize int, idleTimeout int) error {
	return nil
}

func (tstDep *TpDeployment) OpenInstance(zookeeperSize int, mask string) error {
	return nil
}

func (tstDep *TpDeployment) DeleteInstance() error {
	return nil
}

func (tstDep *TpDeployment) StatusInstance() error {
	return nil
}

func (tstDep *TpDeployment) RunClientInstance(cmdArray []string) error {
	return nil
}

func (tstDep *TpDeployment) InstanceExists() bool {
	return tstDep.TstInstanceExists
}

func (tstDep *TpDeployment) FullStatus() (*StardogDescription, error) {
	return tstDep.SdDesc, nil
}
