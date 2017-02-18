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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/stardog-union/stardog-graviton/aws"
	"github.com/stardog-union/stardog-graviton/sdutils"
)

var (
	pluginsMap  map[string]sdutils.Plugin
	consoleFile *os.File
)

// This structure is everything that can come into the CLI
type CliContext struct {
	// Common options
	LicensePath       string                `json:"license_path,omitempty"`
	PrivateKeyPath    string                `json:"private_key,omitempty"`
	LogLevel          string                `json:"log_level,omitempty"`
	CloudType         string                `json:"cloud_type,omitempty"`
	VolumeSize        int                   `json:"volume_size,omitempty"`
	Quiet             bool                  `json:"quiet,omitempty"`
	ClusterSize       int                   `json:"cluster_size,omitempty"`
	SdReleaseFilePath string                `json:"release_file,omitempty"`
	ZkClusterSize     int                   `json:"zookeeper_size,omitempty"`
	Version           string                `json:"sd_version,omitempty"`
	CustomSdProps     string                `json:"custom_stardog_properties,omitempty"`
	OutputFile        string                `json:"output_file,omitempty"`
	HTTPMask          string                `json:"http_mask,omitempty"`
	CloudOpts         interface{}           `json:"cloud_options"`
	DeploymentName    string                `json:"-"`
	CommandList       []string              `json:"-"`
	ConfigDir         string                `json:"-"`
	LogFilePath       string                `json:"-"`
	VerboseLevel      int                   `json:"-"`
	ConsoleLevel      int                   `json:"-"`
	Logger            sdutils.SdVaLogger    `json:"-"`
	InternalHealth    bool                  `json:"-"`
	Force             bool                  `json:"-"`
	Destroy           bool                  `json:"-"`
	NoWaitForHealthy  bool                  `json:"-"`
	WaitMaxTimeSec    int                   `json:"-"`
	ConsoleFile       string                `json:"-"`
	ConsoleWriter     io.Writer             `json:"-"`
	highlight         sdutils.ConsoleEffect `json:"-"`
	red               sdutils.ConsoleEffect `json:"-"`
	green             sdutils.ConsoleEffect `json:"-"`
}

func main() {
	rc := realMain(os.Args[1:])
	os.Exit(rc)
}

func realMain(args []string) int {
	pluginsMap = make(map[string]sdutils.Plugin)
	awsPlugin := aws.GetPlugin()
	pluginsMap[awsPlugin.GetName()] = awsPlugin

	app, err := parseParameters(args)
	if consoleFile != nil {
		consoleFile.Close()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", app.FailString("Failed:"), err.Error())
		return 1
	}
	app.ConsoleLog(1, app.SuccessString("Success.\n"))
	return 0
}

func (cliContext *CliContext) SSHIn(c *kingpin.ParseContext) error {
	baseD := sdutils.BaseDeployment{
		Name:            cliContext.DeploymentName,
		Version:         cliContext.Version,
		Type:            strings.ToLower(cliContext.CloudType),
		Directory:       sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName),
		PrivateKey:      cliContext.PrivateKeyPath,
		CustomPropsFile: cliContext.CustomSdProps,
	}
	d, err := sdutils.LoadDeployment(cliContext, &baseD, false)
	if err != nil {
		return err
	}
	return sdutils.RunSSH(cliContext, &baseD, d)
}

func (cliContext *CliContext) AboutCommand(c *kingpin.ParseContext) error {
	v, err := Asset("etc/version")
	if err != nil {
		return err
	}
	cliContext.ConsoleLog(0, "\n")
	cliContext.ConsoleLog(0, "              Stardog Graviton\n")
	cliContext.ConsoleLog(0, "              Version %s\n", cliContext.GetVersion())
	cliContext.ConsoleLog(0, "              Git hash %s\n", string(v))
	colorDog, err := Asset("etc/stardog.bb")
	sdutils.BbCode(string(colorDog))
	cliContext.ConsoleLog(1, "For a quick start run:\n")
	cliContext.ConsoleLog(1, "%s launch mystardog\n", os.Args[0])

	return nil
}

func (cliContext *CliContext) Interactive(c *kingpin.ParseContext) error {
	var err error

	plugin, err := sdutils.GetPlugin(cliContext.CloudType)
	if err != nil {
		return err
	}

	err = sdutils.AskUserInteractiveString("What version of stardog are you launching?", cliContext.Version, !cliContext.Force, &cliContext.Version)
	if err != nil {
		return err
	}

	if !plugin.HaveImage(cliContext) {
		err = sdutils.AskUserInteractiveString("What is the path to the Stardog release?", cliContext.SdReleaseFilePath, !cliContext.Force, &cliContext.SdReleaseFilePath)
		if err != nil {
			return err
		}
		cliContext.ConsoleLog(0, "There is no base image for version %s.\n", cliContext.Version)
		if !cliContext.Force || sdutils.AskUserYesOrNo("Do you wish to build one?") {
			err = plugin.BuildImage(cliContext, cliContext.SdReleaseFilePath, cliContext.Version)
			if err != nil {
				cliContext.ConsoleLog(0, "Failed to make the stardog base image: %s\n", err.Error())
				return err
			}
		} else {
			return fmt.Errorf("A base image is needed in order to launch a stardog-graviton cluster")
		}
	}
	err = sdutils.AskUserInteractiveString("What would you like to name this deployment?", cliContext.DeploymentName, !cliContext.Force, &cliContext.DeploymentName)
	if err != nil {
		return err
	}

	err = sdutils.AskUserInteractiveString("What CIDR will be allowed to access stardog?", cliContext.HTTPMask, !cliContext.Force, &cliContext.HTTPMask)
	if err != nil {
		return err
	}
	baseD := sdutils.BaseDeployment{
		Name:            cliContext.DeploymentName,
		Version:         cliContext.Version,
		Type:            strings.ToLower(cliContext.CloudType),
		Directory:       sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName),
		PrivateKey:      cliContext.PrivateKeyPath,
		CustomPropsFile: cliContext.CustomSdProps,
	}

	dep, err := sdutils.LoadDeployment(cliContext, &baseD, false)
	if err != nil {
		cliContext.ConsoleLog(1, "Creating the new deployment %s\n", cliContext.DeploymentName)
		dep, err = LoadDepWrapper(cliContext, true)
		if err != nil {
			return err
		}
	}
	if !dep.VolumeExists() {
		err = sdutils.AskUserInteractiveString("What is the path to your Stardog license?", cliContext.LicensePath, !cliContext.Force, &cliContext.LicensePath)
		if err != nil {
			return err
		}
		err = sdutils.AskUserInteractiveInt("How big should each disk be in gigabytes?", cliContext.VolumeSize, !cliContext.Force, &cliContext.VolumeSize)
		if err != nil {
			return err
		}
		err = sdutils.AskUserInteractiveInt("How many Stardog nodes will be in the cluster?", cliContext.ClusterSize, !cliContext.Force, &cliContext.ClusterSize)
		if err != nil {
			return err
		}
		err = dep.CreateVolumeSet(cliContext.LicensePath, cliContext.VolumeSize, cliContext.ClusterSize)
		if err != nil {
			return err
		}
	}
	err = sdutils.AskUserInteractiveInt("How many Zookeeper nodes will be used?", cliContext.ZkClusterSize, !cliContext.Force, &cliContext.ZkClusterSize)
	if err != nil {
		return err
	}
	err = sdutils.CreateInstance(cliContext, &baseD, dep, cliContext.ZkClusterSize, cliContext.WaitMaxTimeSec, cliContext.HTTPMask, cliContext.NoWaitForHealthy)
	if err != nil {
		return err
	}
	return sdutils.FullStatus(cliContext, &baseD, dep, false, cliContext.OutputFile)
}

func (cliContext *CliContext) BaseAmiAction(c *kingpin.ParseContext) error {
	p, err := sdutils.GetPlugin(cliContext.CloudType)
	if err != nil {
		return err
	}
	err = p.BuildImage(cliContext, cliContext.SdReleaseFilePath, cliContext.Version)
	if err != nil {
		cliContext.ConsoleLog(0, "Failed to make the stardog base image: %s\n", err.Error())
		return err
	}
	cliContext.ConsoleLog(1, "The image was successfully created.\n")
	return nil
}

func (cliContext *CliContext) Leaks(c *kingpin.ParseContext) error {
	p, err := sdutils.GetPlugin(cliContext.CloudType)
	if err != nil {
		return err
	}
	return p.FindLeaks(cliContext, cliContext.DeploymentName, cliContext.Destroy, cliContext.Force)
}

func (cliContext *CliContext) NewDeployment(c *kingpin.ParseContext) error {
	// In the future we will check instance types
	_, err := LoadDepWrapper(cliContext, true)
	return err
}

func (cliContext *CliContext) DeploymentList(c *kingpin.ParseContext) error {
	depDir := sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName)
	files, _ := ioutil.ReadDir(depDir)
	for _, f := range files {
		if f.IsDir() {
			cliContext.ConsoleLog(0, "%s\n", f.Name())
		}
	}
	return nil
}

func (cliContext *CliContext) DestroyDeployment(c *kingpin.ParseContext) error {
	if !cliContext.Force && !sdutils.AskUserYesOrNo("Do you really want to destroy?") {
		return nil
	}
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	if d.InstanceExists() {
		return fmt.Errorf("An instance exist in this deployment.  Please destroy it before removing the deployment")
	}
	if d.VolumeExists() {
		return fmt.Errorf("Volumes exist in this deployment.  Please destroy them before removing the deployment")
	}
	cliContext.Force = true
	err = cliContext.DestroyInstance(c)
	if err != nil {
		cliContext.ConsoleLog(1, "The instance was not destroyed %s\n", err)
	}
	err = cliContext.DestroyVolumes(c)
	if err != nil {
		cliContext.ConsoleLog(1, "The volumes were not destroyed %s\n", err)
	}
	sdutils.DeleteDeployment(cliContext, cliContext.DeploymentName)
	return nil
}

func (cliContext *CliContext) FullStatus(c *kingpin.ParseContext) error {
	baseD := sdutils.BaseDeployment{
		Name:            cliContext.DeploymentName,
		Version:         cliContext.Version,
		Type:            strings.ToLower(cliContext.CloudType),
		Directory:       sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName),
		PrivateKey:      cliContext.PrivateKeyPath,
		CustomPropsFile: cliContext.CustomSdProps,
	}
	d, err := sdutils.LoadDeployment(cliContext, &baseD, false)
	if err != nil {
		return err
	}

	return sdutils.FullStatus(cliContext, &baseD, d, cliContext.InternalHealth, cliContext.OutputFile)
}

func (cliContext *CliContext) DestroyFullDeployment(c *kingpin.ParseContext) error {
	cliContext.ConsoleLog(0, "This will destroy all volumes and instances associated with this deployment.\n")
	if !cliContext.Force && !sdutils.AskUserYesOrNo("Do you really want to destroy?") {
		return nil
	}
	cliContext.Force = true
	err := cliContext.DestroyInstance(c)
	if err != nil {
		cliContext.ConsoleLog(1, "The instance was not destroyed.  %s\n", err)
	}
	err = cliContext.DestroyVolumes(c)
	if err != nil {
		cliContext.ConsoleLog(1, "The volumes were not destroyed.  %s\n", err)
	}
	sdutils.DeleteDeployment(cliContext, cliContext.DeploymentName)
	return nil
}

func (cliContext *CliContext) NewVolumes(c *kingpin.ParseContext) error {
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	return d.CreateVolumeSet(cliContext.LicensePath, cliContext.VolumeSize, cliContext.ClusterSize)
}

func (cliContext *CliContext) DestroyVolumes(c *kingpin.ParseContext) error {
	if !cliContext.Force && !sdutils.AskUserYesOrNo("Do you really want to destroy?") {
		return nil
	}
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	return d.DeleteVolumeSet()
}

func (cliContext *CliContext) StatusVolumes(c *kingpin.ParseContext) error {
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	return d.StatusVolumeSet()
}

func (cliContext *CliContext) LaunchInstance(c *kingpin.ParseContext) error {
	baseD := sdutils.BaseDeployment{
		Name:            cliContext.DeploymentName,
		Version:         cliContext.Version,
		Type:            strings.ToLower(cliContext.CloudType),
		Directory:       sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName),
		PrivateKey:      cliContext.PrivateKeyPath,
		CustomPropsFile: cliContext.CustomSdProps,
	}
	dep, err := sdutils.LoadDeployment(cliContext, &baseD, false)
	if err != nil {
		return err
	}

	return sdutils.CreateInstance(cliContext, &baseD, dep, cliContext.ZkClusterSize, cliContext.WaitMaxTimeSec, cliContext.HTTPMask, cliContext.NoWaitForHealthy)
}

func (cliContext *CliContext) DestroyInstance(c *kingpin.ParseContext) error {
	if !cliContext.Force && !sdutils.AskUserYesOrNo("Do you really want to destroy?") {
		return nil
	}
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	return d.DeleteInstance()
}

func (cliContext *CliContext) StatusInstance(c *kingpin.ParseContext) error {
	d, err := LoadDepWrapper(cliContext, false)
	if err != nil {
		return err
	}
	return d.StatusInstance()
}

func (cliContext *CliContext) ConsoleLog(level int, format string, v ...interface{}) {
	if level > cliContext.ConsoleLevel {
		return
	}
	_, err := fmt.Fprintf(cliContext.ConsoleWriter, format, v...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to the file %s\n", err)
	}
}

func (cliContext *CliContext) HighlightString(a ...interface{}) string {
	return cliContext.highlight(a...)
}

func (cliContext *CliContext) SuccessString(a ...interface{}) string {
	return cliContext.green(a...)
}

func (cliContext *CliContext) FailString(a ...interface{}) string {
	return cliContext.red(a...)
}

func (cliContext *CliContext) GetConfigDir() string {
	return cliContext.ConfigDir
}

func (cliContext *CliContext) GetVersion() string {
	return cliContext.Version
}

func (cliContext *CliContext) Logf(level int, format string, v ...interface{}) {
	cliContext.Logger.Logf(level, format, v...)
}

func (cliContext *CliContext) nameValidate(a *kingpin.CmdClause) error {
	if len(cliContext.DeploymentName) > 20 {
		return fmt.Errorf("Could the deployment name must be less than 20 characters")
	}
	return nil
}

func (cliContext *CliContext) topValidate(a *kingpin.Application) error {
	var err error
	// Normalize the options
	if _, err = os.Stat(cliContext.ConfigDir); os.IsNotExist(err) {
		os.MkdirAll(cliContext.ConfigDir, 0755)
	}

	if cliContext.LogFilePath == "" {
		if cliContext.DeploymentName != "" {
			cliContext.LogFilePath = filepath.Join(sdutils.DeploymentDir(cliContext.ConfigDir, cliContext.DeploymentName), "logs", "graviton.log")
		} else {
			cliContext.LogFilePath = filepath.Join(cliContext.ConfigDir, "logs", "graviton.log")
		}
	}
	if cliContext.ConsoleFile == "" {
		cliContext.ConsoleWriter = os.Stdout
	} else {
		consoleFile, err = os.OpenFile(cliContext.ConsoleFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Failed to create the console file %s", cliContext.ConsoleFile)
		}
		cliContext.ConsoleWriter = consoleFile
	}

	logDir := path.Dir(cliContext.LogFilePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}
	logWriter, err := os.OpenFile(cliContext.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Could not open the file %s.  %s", cliContext.LogFilePath, err.Error())
	}
	cliContext.LogLevel = strings.ToUpper(cliContext.LogLevel)
	baseLogger := log.New(logWriter, "", log.Ldate|log.Ltime)
	cliContext.Logger, err = sdutils.NewSdVaLogger(baseLogger, cliContext.LogLevel)

	cliContext.ConsoleLevel = cliContext.VerboseLevel + 1
	if cliContext.Quiet {
		cliContext.ConsoleLevel = 0
	}
	cliContext.ConsoleLog(2, "Logging to the file %s\n", cliContext.LogFilePath)
	return nil
}

func loadDefaultCliOptions() *CliContext {
	usr, _ := user.Current()
	confDir := os.Getenv("STARDOG_VIRTUAL_APPLIANCE_CONFIG_DIR")
	if confDir == "" {
		confDir = filepath.Join(usr.HomeDir, ".graviton")
	}
	// Setup defaults here
	cliContext := CliContext{
		DeploymentName:    "",
		ConfigDir:         confDir,
		LogLevel:          "INFO",
		CloudType:         "aws",
		VolumeSize:        10,
		Quiet:             false,
		ClusterSize:       3,
		SdReleaseFilePath: "",
		ZkClusterSize:     3,
		Version:           "",
		LicensePath:       "",
		PrivateKeyPath:    "",
		Force:             false,
		NoWaitForHealthy:  false,
		WaitMaxTimeSec:    600,
		HTTPMask:          sdutils.GetLocalOnlyHTTPMask(),
		highlight:         color.New(color.FgHiWhite, color.Bold).SprintFunc(),
		green:             color.New(color.FgGreen, color.Bold).SprintFunc(),
		red:               color.New(color.FgRed, color.Bold).SprintFunc(),
	}
	defaultFile := filepath.Join(confDir, "default.json")

	if os.Getenv("STARDOG_GRAVITON_UNIT_TEST") == "" {
		sdutils.LoadJSON(&cliContext, defaultFile)
	}
	p, ok := pluginsMap[cliContext.CloudType]
	if ok && cliContext.CloudType != "" {
		err := p.LoadDefaults(cliContext.CloudOpts)
		if err != nil {
			fmt.Printf("Failed to load the default cloud opts: %s\n", err)
		}
	}

	return &cliContext
}

func parseParameters(args []string) (*CliContext, error) {
	// Setup defaults here
	cliContext := loadDefaultCliOptions()

	cmdOpts := sdutils.CommandOpts{}
	cli := kingpin.New("stardog-graviton", "The stardog virtual appliance manager.")
	cmdOpts.Cli = cli

	versionString := "unknown"
	b, err := Asset("etc/version")
	if err == nil {
		versionString = strings.TrimSpace(string(b))
	}
	cli.Version(versionString)

	cli.Flag("console-file", "Instead of sending console output to stdout send it here").StringVar(&cliContext.ConsoleFile)
	cli.Flag("log-level", fmt.Sprintf("Log level [%s]", strings.Join(sdutils.LogLevelNames, " | "))).Default(cliContext.LogLevel).StringVar(&cliContext.LogLevel)
	cli.Flag("config-dir", "The path for the log file").Default(cliContext.ConfigDir).StringVar(&cliContext.ConfigDir)
	cli.Flag("verbose", "How much output to send to the console").CounterVar(&cliContext.VerboseLevel)
	cli.Flag("quiet", "Minimal console output").Default(fmt.Sprintf("%t", cliContext.Quiet)).BoolVar(&cliContext.Quiet)
	cli.Validate(cliContext.topValidate)

	cmdOpts.LaunchCmd = cli.Command("launch", "Walk through a launch from scratch.")
	cmdOpts.LaunchCmd.Flag("interactive", "Ask all questions even if there are default values.").Default(fmt.Sprintf("%t", cliContext.Force)).BoolVar(&cliContext.Force)
	cmdOpts.LaunchCmd.Flag("type", "The type of cloud with which graviton will interact (only aws supported).").Default(cliContext.CloudType).StringVar(&cliContext.CloudType)
	cmdOpts.LaunchCmd.Flag("name", "The name of the deployment.  It must be unique to this account.").StringVar(&cliContext.DeploymentName)
	cmdOpts.LaunchCmd.Flag("sd-version", "The stardog version to associate with this deployment.").Default(cliContext.Version).StringVar(&cliContext.Version)
	cmdOpts.LaunchCmd.Flag("private-key", "The path to the private key").Default(cliContext.PrivateKeyPath).StringVar(&cliContext.PrivateKeyPath)
	cmdOpts.LaunchCmd.Flag("license", "Path to your stardog license.").Default(cliContext.LicensePath).StringVar(&cliContext.LicensePath)
	cmdOpts.LaunchCmd.Flag("release", "Path to the stardog release zip file.").Default(cliContext.SdReleaseFilePath).StringVar(&cliContext.SdReleaseFilePath)
	cmdOpts.LaunchCmd.Flag("volume-size", "The size of each storage volume in gigabytes.").Default(fmt.Sprintf("%d", cliContext.VolumeSize)).IntVar(&cliContext.VolumeSize)
	cmdOpts.LaunchCmd.Flag("node-count", "The number storage volume.  This will be the size of the stardog cluster").Default(fmt.Sprintf("%d", cliContext.ClusterSize)).IntVar(&cliContext.ClusterSize)
	cmdOpts.LaunchCmd.Flag("zk-count", "The number of zookeeper nodes.").Default(fmt.Sprintf("%d", cliContext.ZkClusterSize)).IntVar(&cliContext.ZkClusterSize)
	cmdOpts.LaunchCmd.Flag("stardog-properties", "A custom stardog properties file.").Default(cliContext.CustomSdProps).StringVar(&cliContext.CustomSdProps)
	cmdOpts.LaunchCmd.Flag("no-wait", "Block until the stardog instance is healthy.").Default(fmt.Sprintf("%t", cliContext.NoWaitForHealthy)).BoolVar(&cliContext.NoWaitForHealthy)
	cmdOpts.LaunchCmd.Flag("wait-timeout", "The number of seconds to block waiting for the stardog instance to become healthy.").Default(fmt.Sprintf("%d", cliContext.WaitMaxTimeSec)).IntVar(&cliContext.WaitMaxTimeSec)
	cmdOpts.LaunchCmd.Arg("name", "The name of the deployment.  It must be unique to this account.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.LaunchCmd.Flag("cidr", "The network mask to which stardog access will be limited.  The default is the IP of this machine.").Default(cliContext.HTTPMask).StringVar(&cliContext.HTTPMask)
	cmdOpts.LaunchCmd.Action(cliContext.Interactive)
	cmdOpts.LaunchCmd.Validate(cliContext.nameValidate)

	cmdOpts.DestroyCmd = cli.Command("destroy", "Destroy everything associated with a deployment.")
	cmdOpts.DestroyCmd.Arg("name", "The name of the deployment to destroy.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.DestroyCmd.Flag("force", "Do not verify with the destruction.").Default("false").BoolVar(&cliContext.Force)
	cmdOpts.DestroyCmd.Action(cliContext.DestroyFullDeployment)

	cmdOpts.StatusCmd = cli.Command("status", "Check the status of a full deployment.")
	cmdOpts.StatusCmd.Arg("deployment name", "The name of the deployment to inspect.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.StatusCmd.Flag("json-file", "The path json output file.").StringVar(&cliContext.OutputFile)
	cmdOpts.StatusCmd.Flag("internal-health", "Do not verify with the destruction.").Default("false").BoolVar(&cliContext.InternalHealth)
	cmdOpts.StatusCmd.Action(cliContext.FullStatus)

	cmdOpts.LeaksCmd = cli.Command("leaks", "Check aws services for possible resource leaks.")
	cmdOpts.LeaksCmd.Flag("destroy", "Destroy any of the resources found.").Default("false").BoolVar(&cliContext.Destroy)
	cmdOpts.LeaksCmd.Flag("force", "Destroy any of the resources found without first asking.").Default("false").BoolVar(&cliContext.Force)
	cmdOpts.LeaksCmd.Flag("deployment-name", "Limit the search to a particular deployment name.").StringVar(&cliContext.DeploymentName)
	cmdOpts.LeaksCmd.Action(cliContext.Leaks)

	cmdOpts.SshCmd = cli.Command("ssh", "ssh into the bastion node.")
	cmdOpts.SshCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.SshCmd.Action(cliContext.SSHIn)

	cmdOpts.AboutCmd = cli.Command("about", "Display information about this program.")
	cmdOpts.AboutCmd.Action(cliContext.AboutCommand)

	cmdOpts.BuildCmd = cli.Command("baseami", "Create a base ami.")
	cmdOpts.BuildCmd.Arg("release", "The stardog release file.").Required().StringVar(&cliContext.SdReleaseFilePath)
	cmdOpts.BuildCmd.Arg("sd-version", "The stardog release version to will be baked into this file.").Required().StringVar(&cliContext.Version)
	cmdOpts.BuildCmd.Action(cliContext.BaseAmiAction)

	deployCmd := cli.Command("deployment", "Manage and inspect deployments.")
	cmdOpts.NewDeploymentCmd = deployCmd.Command("new", "Define a new deployment but do not create volumes or launch an instance.")
	cmdOpts.NewDeploymentCmd.Flag("type", "The type of cloud with which graviton will interact (only aws supported).").Default("aws").StringVar(&cliContext.CloudType)
	cmdOpts.NewDeploymentCmd.Arg("name", "The name of the deployment.  It must be unique to this account.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.NewDeploymentCmd.Arg("sd-version", "The stardog version to associate with this deployment.").Required().StringVar(&cliContext.Version)
	cmdOpts.NewDeploymentCmd.Flag("private-key", "The path to the private key.").Default(cliContext.PrivateKeyPath).StringVar(&cliContext.PrivateKeyPath)
	cmdOpts.NewDeploymentCmd.Flag("stardog-properties", "A custom stardog properties file.").Default(cliContext.CustomSdProps).StringVar(&cliContext.CustomSdProps)
	cmdOpts.NewDeploymentCmd.Action(cliContext.NewDeployment)
	cmdOpts.NewDeploymentCmd.Validate(cliContext.nameValidate)

	cmdOpts.DestroyDeploymentCmd = deployCmd.Command("destroy", "Destroy the deployment.  This will fail if volumes exist or an instance is running.")
	cmdOpts.DestroyDeploymentCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.DestroyDeploymentCmd.Flag("force", "Do not verify with the destruction.").Default("false").BoolVar(&cliContext.Force)
	cmdOpts.DestroyDeploymentCmd.Action(cliContext.DestroyDeployment)

	cmdOpts.ListDeploymentCmd = deployCmd.Command("list", "List the knwon deployments.")
	cmdOpts.ListDeploymentCmd.Action(cliContext.DeploymentList)

	volumesCmd := cli.Command("volume", "Manage storage volumes.")
	cmdOpts.NewVolumesCmd = volumesCmd.Command("new", "Create new backing storage.")
	cmdOpts.NewVolumesCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.NewVolumesCmd.Arg("license", "Path to your stardog license.").Required().StringVar(&cliContext.LicensePath)
	cmdOpts.NewVolumesCmd.Arg("size", "The size of each storage volume in gigabytes.").Required().IntVar(&cliContext.VolumeSize)
	cmdOpts.NewVolumesCmd.Arg("count", "The number storage volume.  This will be the size of the stardog cluster.").Required().IntVar(&cliContext.ClusterSize)
	cmdOpts.NewVolumesCmd.Action(cliContext.NewVolumes)

	cmdOpts.DestroyVolumesCmd = volumesCmd.Command("destroy", "This will destroy the volumes permanently.")
	cmdOpts.DestroyVolumesCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.DestroyVolumesCmd.Flag("force", "Do not verify with the destruction.").Default("false").BoolVar(&cliContext.Force)
	cmdOpts.DestroyVolumesCmd.Action(cliContext.DestroyVolumes)

	cmdOpts.StatusVolumesCmd = volumesCmd.Command("status", "Display information about the volumes.")
	cmdOpts.StatusVolumesCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.StatusVolumesCmd.Action(cliContext.StatusVolumes)

	instanceCmd := cli.Command("instance", "Manage the instance.")
	cmdOpts.LaunchInstanceCmd = instanceCmd.Command("new", "Create new set of VMs running Stardog.")
	cmdOpts.LaunchInstanceCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.LaunchInstanceCmd.Arg("zk", "The number of zookeeper nodes.").Required().IntVar(&cliContext.ZkClusterSize)
	cmdOpts.LaunchInstanceCmd.Flag("no-wait", "Block until the stardog instance is healthy.").Default(fmt.Sprintf("%t", cliContext.NoWaitForHealthy)).BoolVar(&cliContext.NoWaitForHealthy)
	cmdOpts.LaunchInstanceCmd.Flag("wait-timeout", "The number of seconds to block waiting for the stardog instance to become healthy.").Default(fmt.Sprintf("%d", cliContext.WaitMaxTimeSec)).IntVar(&cliContext.WaitMaxTimeSec)
	cmdOpts.LaunchInstanceCmd.Flag("cidr", "The network mask to which stardog access will be limited.").StringVar(&cliContext.HTTPMask)
	cmdOpts.LaunchInstanceCmd.Action(cliContext.LaunchInstance)

	cmdOpts.DestroyInstanceCmd = instanceCmd.Command("destroy", "Destroy the instance.")
	cmdOpts.DestroyInstanceCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.DestroyInstanceCmd.Flag("force", "Do not verify with the destruction.").Default("false").BoolVar(&cliContext.Force)
	cmdOpts.DestroyInstanceCmd.Action(cliContext.DestroyInstance)

	cmdOpts.StatusInstanceCmd = instanceCmd.Command("status", "Get information about the instance.")
	cmdOpts.StatusInstanceCmd.Arg("deployment", "The name of the deployment.").Required().StringVar(&cliContext.DeploymentName)
	cmdOpts.StatusInstanceCmd.Action(cliContext.StatusInstance)

	// Add all the options for all the plugins
	for _, p := range pluginsMap {
		p.Register(&cmdOpts)
		sdutils.AddCloudType(p)
	}

	_, err = cli.Parse(args)
	if err != nil {
		return cliContext, err
	}

	return cliContext, nil
}

func LoadDepWrapper(cliContext *CliContext, new bool) (sdutils.Deployment, error) {
	baseD := sdutils.BaseDeployment{
		Name:            cliContext.DeploymentName,
		Version:         cliContext.Version,
		Type:            strings.ToLower(cliContext.CloudType),
		Directory:       sdutils.DeploymentDir(cliContext.GetConfigDir(), cliContext.DeploymentName),
		PrivateKey:      cliContext.PrivateKeyPath,
		CustomPropsFile: cliContext.CustomSdProps,
	}
	return sdutils.LoadDeployment(cliContext, &baseD, new)
}
