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
	"gopkg.in/alecthomas/kingpin.v2"
)

type ConsoleEffect func(a ...interface{}) string

type BaseDeployment struct {
	Type            string      `json:"type,omitempty"`
	Name            string      `json:"name,omitempty"`
	Directory       string      `json:"directory,omitempty"`
	Version         string      `json:"version,omitempty"`
	PrivateKey      string      `json:"private_key,omitempty"`
	CustomPropsFile string      `json:"custom_props,omitempty"`
	IdleTimeout     int         `json:"idle_timeout,omitempty"`
	CloudOpts       interface{} `json:"cloud_opts,omitempty"`
}

type AppContext interface {
	ConsoleLog(level int, format string, v ...interface{})
	Logf(level int, format string, v ...interface{})
	GetConfigDir() string
	GetVersion() string
	HighlightString(a ...interface{}) string
	SuccessString(a ...interface{}) string
	FailString(a ...interface{}) string
}

type StardogDescription struct {
	StardogURL          string      `json:"stardog_url,omitempty"`
	StardogInternalURL  string      `json:"stardog_internal_url,omitempty"`
	SSHHost             string      `json:"ssh_host,omitempty"`
	Healthy             bool        `json:"healthy,omitempty"`
	VolumeDescription   interface{} `json:"volume,omitempty"`
	InstanceDescription interface{} `json:"instance,omitempty"`
}

type Deployment interface {
	CreateVolumeSet(licensePath string, sizeOfEachVolume int, clusterSize int) error
	DeleteVolumeSet() error
	StatusVolumeSet() error
	VolumeExists() bool

	CreateInstance(zookeeperSize int, idleTimeout int) error
	OpenInstance(zookeeperSize int, mask string, idleTimeout int) error
	DeleteInstance() error
	StatusInstance() error
	InstanceExists() bool

	FullStatus() (*StardogDescription, error)
}

type CommandOpts struct {
	Cli                  *kingpin.Application
	LaunchCmd            *kingpin.CmdClause
	DestroyCmd           *kingpin.CmdClause
	StatusCmd            *kingpin.CmdClause
	LeaksCmd             *kingpin.CmdClause
	ClientCmd            *kingpin.CmdClause
	SshCmd               *kingpin.CmdClause
	PasswdCmd            *kingpin.CmdClause
	AboutCmd             *kingpin.CmdClause
	BuildCmd             *kingpin.CmdClause
	NewDeploymentCmd     *kingpin.CmdClause
	DestroyDeploymentCmd *kingpin.CmdClause
	ListDeploymentCmd    *kingpin.CmdClause
	NewVolumesCmd        *kingpin.CmdClause
	DestroyVolumesCmd    *kingpin.CmdClause
	StatusVolumesCmd     *kingpin.CmdClause
	LaunchInstanceCmd    *kingpin.CmdClause
	DestroyInstanceCmd   *kingpin.CmdClause
	StatusInstanceCmd    *kingpin.CmdClause
}

type Plugin interface {
	Register(cmdOpts *CommandOpts) error
	DeploymentLoader(context AppContext, baseD *BaseDeployment, new bool) (Deployment, error)
	LoadDefaults(defaultCliOpts interface{}) error
	BuildImage(context AppContext, sdReleaseFilePath string, version string) error
	GetName() string
	FindLeaks(context AppContext, deploymentName string, destroy bool, force bool) error
	HaveImage(context AppContext) bool
}
