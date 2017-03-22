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

// TestContext is used for mocking out a context in many tests
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

type tstPlugin struct {
	HasImage bool
	Dep      Deployment
}

func (tp *tstPlugin) Register(cmdOpts *CommandOpts) error {
	return nil
}

func (tp *tstPlugin) DeploymentLoader(context AppContext, baseD *BaseDeployment, new bool) (Deployment, error) {
	if tp.Dep == nil {
		return nil, fmt.Errorf("This test plugin loads no deployment")
	}
	return tp.Dep, nil
}

func (tp *tstPlugin) LoadDefaults(defaultCliOpts interface{}) error {
	return nil
}

func (tp *tstPlugin) BuildImage(context AppContext, sdReleaseFilePath string, version string) error {
	return nil
}

func (tp *tstPlugin) GetName() string {
	return "test1loaderror"
}

func (tp *tstPlugin) FindLeaks(context AppContext, deploymentName string, destroy bool, force bool) error {
	return nil
}

func (tp *tstPlugin) HaveImage(context AppContext) bool {
	return tp.HasImage
}

type tpDeployment struct {
	TstInstanceExists bool
	TstVolumeExists   bool
	SdDesc            *StardogDescription
}

func (tstDep *tpDeployment) CreateVolumeSet(licensePath string, sizeOfEachVolume int, clusterSize int) error {
	return nil
}

func (tstDep *tpDeployment) DeleteVolumeSet() error {
	return nil
}

func (tstDep *tpDeployment) StatusVolumeSet() error {
	return nil
}

func (tstDep *tpDeployment) VolumeExists() bool {
	return tstDep.TstVolumeExists
}

func (tstDep *tpDeployment) CreateInstance(zookeeperSize int, idleTimeout int) error {
	return nil
}

func (tstDep *tpDeployment) OpenInstance(zookeeperSize int, mask string, idleTimeout int) error {
	return nil
}

func (tstDep *tpDeployment) DeleteInstance() error {
	return nil
}

func (tstDep *tpDeployment) StatusInstance() error {
	return nil
}

func (tstDep *tpDeployment) RunClientInstance(cmdArray []string) error {
	return nil
}

func (tstDep *tpDeployment) InstanceExists() bool {
	return tstDep.TstInstanceExists
}

func (tstDep *tpDeployment) FullStatus() (*StardogDescription, error) {
	return tstDep.SdDesc, nil
}

func (tstDep *tpDeployment) ClusterSize() (int, error) {
	return 1, nil
}

func (tstDep *tpDeployment) DestroyDeployment() error {
	return nil
}

