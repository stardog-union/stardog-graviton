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
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

// ConsoleEffect is a function for writing lines to the console in a
// visually pleasing way.  For example red text for error messages.
type ConsoleEffect func(a ...interface{}) string

// BaseDeployment hold information about the deployments and is serialized
// to JSON.  CloudOpts is defined by the specific plugin in use.
type BaseDeployment struct {
	Type            string      `json:"type,omitempty"`
	Name            string      `json:"name,omitempty"`
	Directory       string      `json:"directory,omitempty"`
	Version         string      `json:"version,omitempty"`
	PrivateKey      string      `json:"private_key,omitempty"`
	CustomPropsFile string      `json:"custom_props,omitempty"`
	CustomLog4J     string      `json:"custom_log4j,omitempty"`
	IdleTimeout     int         `json:"idle_timeout,omitempty"`
	Environment     []string    `json:"environment,omitempty"`
	DisableSecurity bool        `json:"disable_security,omitempty"`
	CloudOpts       interface{} `json:"cloud_opts,omitempty"`
	CustomScript    string      `json:"custom_script,omitempty"`
	CustomZkScript  string      `json:"custom_zk_script,omitempty"`
}

// AppContext provides and abstraction to logging, console interaction and
// basic configuration information
type AppContext interface {
	ConsoleLog(level int, format string, v ...interface{})
	Logf(level int, format string, v ...interface{})
	GetConfigDir() string
	GetVersion() string
	GetInteractive() bool
	HighlightString(a ...interface{}) string
	SuccessString(a ...interface{}) string
	FailString(a ...interface{}) string
}

// StardogDescription represents the state of a Stardog deployment.  It is effectively
// the output from a status command.
type StardogDescription struct {
	StardogURL          string      `json:"stardog_url,omitempty"`
	StardogInternalURL  string      `json:"stardog_internal_url,omitempty"`
	StardogNodes        []string    `json:"stardog_nodes,omitempty"`
	SSHHost             string      `json:"ssh_host,omitempty"`
	Healthy             bool        `json:"healthy,omitempty"`
	TimeStamp           time.Time   `json:"timestamp,omitempty"`
	VolumeDescription   interface{} `json:"volume,omitempty"`
	InstanceDescription interface{} `json:"instance,omitempty"`
}

// Deployment is an interface to a plugin that is managing the actual Stardog services.
type Deployment interface {
	CreateVolumeSet(licensePath string, sizeOfEachVolume int, clusterSize int) error
	DeleteVolumeSet() error
	StatusVolumeSet() error
	VolumeExists() bool
	ClusterSize() (int, error)

	CreateInstance(volumeSize int, zookeeperSize int, idleTimeout int) error
	OpenInstance(volumeSize int, zookeeperSize int, mask string, idleTimeout int) error
	DeleteInstance() error
	StatusInstance() error
	InstanceExists() bool

	FullStatus() (*StardogDescription, error)

	DestroyDeployment() error
}

// CommandOpts holds all of the CLI parsing information for the system.
// It is passed to plugins so that each driver can add their own specific
// flags.
type CommandOpts struct {
	Cli                  *kingpin.Application
	LaunchCmd            *kingpin.CmdClause
	DestroyCmd           *kingpin.CmdClause
	StatusCmd            *kingpin.CmdClause
	LeaksCmd             *kingpin.CmdClause
	ClientCmd            *kingpin.CmdClause
	SSHCmd               *kingpin.CmdClause
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

// Plugin defines the interface for adding drivers to the system
type Plugin interface {
	Register(cmdOpts *CommandOpts) error
	DeploymentLoader(context AppContext, baseD *BaseDeployment, new bool) (Deployment, error)
	LoadDefaults(defaultCliOpts interface{}) error
	BuildImage(context AppContext, sdReleaseFilePath string, version string) error
	GetName() string
	FindLeaks(context AppContext, deploymentName string, destroy bool, force bool) error
	HaveImage(context AppContext) bool
}
