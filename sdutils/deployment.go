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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	pluginMap = make(map[string]Plugin)
)

// AddCloudType will associate a new plugin type  with this graviton instances
func AddCloudType(p Plugin) {
	pluginMap[p.GetName()] = p
}

// GetPlugin returns the plugin associate with the given name
func GetPlugin(name string) (Plugin, error) {
	p, ok := pluginMap[name]
	if !ok {
		return nil, fmt.Errorf("The plugin %s does not exist", name)
	}
	return p, nil
}

// DeploymentDir abstracts the location of deployment information files into
// a function
func DeploymentDir(confDir string, deploymentName string) string {
	return path.Join(confDir, "deployments", deploymentName)
}

// DeleteDeployment will remove all information stored on the local file system that
// is associated with a deployment.
func DeleteDeployment(context AppContext, name string) {
	deploymentDir := DeploymentDir(context.GetConfigDir(), name)
	os.RemoveAll(deploymentDir)
}

// LoadDeployment inflates a Deployment object from the information stored in the
// configuration directory.
func LoadDeployment(context AppContext, baseD *BaseDeployment, new bool) (Deployment, error) {
	confPath := path.Join(baseD.Directory, "config.json")

	plugin, err := GetPlugin(baseD.Type)
	if err != nil {
		return nil, err
	}

	if !new {
		if _, err := os.Stat(confPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("The deployment %s does not exist", baseD.Name)
		}
		data, err := ioutil.ReadFile(confPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, baseD)
		if err != nil {
			return nil, err
		}
		context.Logf(DEBUG, "Loading the default %s from %s", baseD, confPath)
		return plugin.DeploymentLoader(context, baseD, new)
	}
	os.MkdirAll(baseD.Directory, 0755)

	if baseD.CustomScript != "" && !PathExists(baseD.CustomScript) {
		return nil, fmt.Errorf("The path to the custom script %s does not exist", baseD.CustomScript)
	}
	if baseD.CustomZkScript != "" && !PathExists(baseD.CustomZkScript) {
		return nil, fmt.Errorf("The path to the custom zk script %s does not exist", baseD.CustomZkScript)
	}

	d, err := plugin.DeploymentLoader(context, baseD, new)
	return d, err
}

func runClient(context AppContext, sd *StardogDescription, baseD *BaseDeployment, d Deployment, cmdArray []string) error {
	baseSSH, err := getSSHCommand(context, baseD, sd)
	if err != nil {
		return nil
	}
	chpwCmd := append(baseSSH,
		"sudo",
		"/usr/local/stardog/bin/stardog-admin",
		"--server",
		sd.StardogInternalURL)
	chpwCmd = append(chpwCmd,
		cmdArray...)

	cmd := exec.Cmd{
		Path: chpwCmd[0],
		Args: chpwCmd,
	}
	_, err = RunCommand(context, cmd, linePrinter, nil)
	return err
}

func getSSHCommand(context AppContext, baseD *BaseDeployment, sd *StardogDescription) ([]string, error) {
	context.Logf(DEBUG, "sshing to %s to run the stardog client\n", sd.SSHHost)

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return nil, err
	}

	sshCmd := []string{sshPath,
		"-t", "-t",
		"-A",
		"-i", baseD.PrivateKey,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", "ubuntu", sd.SSHHost),
	}

	return sshCmd, nil
}

func runSCPCommand(context AppContext, baseD *BaseDeployment, sd *StardogDescription, local string, remote string, upload bool) error {
	context.Logf(DEBUG, "sshing to %s to run the stardog client\n", sd.SSHHost)

	scpPath, err := exec.LookPath("scp")
	if err != nil {
		return err
	}

	var scpStrA []string
	remoteTarget := fmt.Sprintf("%s@%s:%s", "ubuntu", sd.SSHHost, remote)
	if upload {
		scpStrA = []string{scpPath,
			"-v",
			"-i", baseD.PrivateKey,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			local,
			remoteTarget,
		}
	} else {
		scpStrA = []string{scpPath,
			"-v",
			"-i", baseD.PrivateKey,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			remoteTarget,
			local,
		}
	}
	scpCmd := exec.Cmd{
		Path: scpStrA[0],
		Args: scpStrA,
	}
	o, err := scpCmd.CombinedOutput()
	if err != nil {
		context.Logf(ERROR, "scp error: %s.  %s", err, string(o))
		return err
	}
	return nil
}

// RunSSH will start an ssh session on the bastion node
func RunSSH(context AppContext, baseD *BaseDeployment, d Deployment) error {
	sd, err := d.FullStatus()
	if err != nil {
		return err
	}

	baseSSH, err := getSSHCommand(context, baseD, sd)
	if err != nil {
		return nil
	}

	cmd := exec.Cmd{
		Path: baseSSH[0],
		Args: baseSSH,
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// IsHealthy checks the deployment to see if the Stardog service is healthy.  if
// internal is set to true it will test by sshing into the bastion node first.
func IsHealthy(context AppContext, baseD *BaseDeployment, d Deployment, internal bool) bool {
	sd, err := d.FullStatus()
	if err != nil {
		context.Logf(WARN, "Status failure %s", err)
		return false
	}
	if os.Getenv("STARDOG_GRAVITON_UNIT_TEST") != "" {
		h := os.Getenv("STARDOG_GRAVITON_HEALTHY")
		if h != "" {
			b, err := strconv.ParseBool(h)
			if err != nil {
				return false
			}
			return b
		}
		return true
	}

	if internal {
		context.Logf(DEBUG, "Checking health via ssh.")
		sshBase, err := getSSHCommand(context, baseD, sd)
		if err != nil {
			context.Logf(DEBUG, "ssh command error %s", err)
			return false
		}

		sshCmd := append(sshBase, []string{
			"/usr/bin/curl",
			"-s", "-o", "/dev/null",
			"-w", "%{http_code}",
			fmt.Sprintf("%s/admin/healthcheck", sd.StardogInternalURL),
		}...)

		cmd := exec.Cmd{
			Path: sshCmd[0],
			Args: sshCmd,
		}

		context.Logf(INFO, "Running the remote health checker %s.", strings.Join(sshCmd, " "))
		b, err := cmd.Output()
		if err != nil {
			context.Logf(DEBUG, "ssh run error %s", err)
			return false
		}
		return string(b) == "200"
	}
	url := fmt.Sprintf("%s/admin/healthcheck", sd.StardogURL)
	context.Logf(DEBUG, "Checking health at %s.", url)

	response, err := http.Get(url)
	if err != nil {
		context.Logf(DEBUG, "Error getting the health check %s", err)
		return false
	}
	return response.StatusCode == 200
}

// WaitForHealth will block until the deployment is considered healthy or the
// timeout expires.  If internal is true it will ssh into the bastion node
// before checking the health URL.
func WaitForHealth(context AppContext, baseD *BaseDeployment, d Deployment, waitTimeout int, internal bool) error {
	last := ""
	pollInterval := 2
	itCnt := waitTimeout / pollInterval

	var spin *Spinner
	if internal {
		spin = NewSpinner(context, 2, "Waiting for the node to be healthy internally")
	} else {
		spin = NewSpinner(context, 1, "Waiting for external health check to pass")
	}
	for i := 0; !IsHealthy(context, baseD, d, internal); i++ {
		if i >= itCnt {
			return fmt.Errorf("Timed out waiting for the instance to get healthy")
		}
		spin.EchoNext()
		context.ConsoleLog(1, "\r%s", last)
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
	spin.Close()
	context.ConsoleLog(1, "%s\n", context.SuccessString("The instance is healthy"))
	return nil
}

func WaitForNClusterNodes(context AppContext, size int, sdURL string, pw string, waitTimeout int) error {
	var err error
	pollInterval := 2
	itCnt := waitTimeout / pollInterval

	client := stardogClientImpl{
		sdURL:    sdURL,
		logger:   context,
		username: "admin",
		password: pw,
	}
	spinner := NewSpinner(context, 2, "Waiting for the node to be healthy internally")
	nodes := &[]string{}
	for i := 0; len(*nodes) < size; i++ {
		context.ConsoleLog(2, "%d nodes waiting for %d\n", len(*nodes), size)
		if i >= itCnt {
			return fmt.Errorf("Timed out waiting for all the cluster nodes")
		}
		spinner.EchoNext()
		time.Sleep(time.Duration(pollInterval) * time.Second)
		nodes, err = client.GetClusterInfo()
		if err != nil {
			context.Logf(WARN, "Cluster info failed: %s", err)
			nodes = &[]string{}
		}
	}
	spinner.Close()
	context.ConsoleLog(1, "%s\n", context.SuccessString("The instance is healthy"))
	return nil
}

func linePrinter(cliContext AppContext, line string) *ScanResult {
	cliContext.Logf(DEBUG, line)
	cliContext.ConsoleLog(1, "%s\n", line)
	return nil
}

// CreateInstance wraps up the deployment.CreateInstance method and blocks until
// the deployment is considered healthy.  It will then change the password by
// SSHing into the bastion node.  Once that is complete it will open up the
// the firewall.
func CreateInstance(context AppContext, baseD *BaseDeployment, dep Deployment, volumeSize int, zkSize int, waitMaxTimeSec int, timeoutSec int, mask string, bastionVolSnapshotId string, noWait bool) error {
	err := dep.CreateInstance(volumeSize, zkSize, timeoutSec, bastionVolSnapshotId)
	if err != nil {
		return err
	}
	if noWait {
		context.ConsoleLog(1, "Not waiting...\n")
		return nil
	}

	context.ConsoleLog(1, "Waiting for stardog to come up...\n")
	err = WaitForHealth(context, baseD, dep, waitMaxTimeSec, true)
	if err != nil {
		return err
	}
	sd, err := dep.FullStatus()
	if err != nil {
		return err
	}
	pw := "admin"
	newPw := os.Getenv("STARDOG_ADMIN_PASSWORD")
	if newPw != "" {
		context.ConsoleLog(1, "Changing the default password...\n")
		err = runClient(context, sd, baseD, dep, []string{"user", "passwd", "-u", "admin", "-N", newPw, "-p", "admin"})
		if err != nil {
			return err
		}
		pw = newPw
	}
	err = dep.OpenInstance(volumeSize, zkSize, mask, timeoutSec)
	if err != nil {
		return err
	}
	clusterSize, err := dep.ClusterSize()
	if err != nil {
		return err
	}
	err = WaitForNClusterNodes(context, clusterSize, sd.StardogURL, pw, waitMaxTimeSec)
	return err
}

// Upload a new Stardog release zip to the nodes and restart Stardog
func UpdateStardog(context AppContext, baseD *BaseDeployment, dep Deployment, sdReleaseFile string) error {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return fmt.Errorf("ssh-agent needs to be setup to update Stardog binaries")
	}
	context.ConsoleLog(1, "Updating Stardog (this may take a few minutes)...\n")
	sd, err := dep.FullStatus()
	if err != nil {
		return err
	}
	sshBase, err := getSSHCommand(context, baseD, sd)
	if err != nil {
		return err
	}
	context.Logf(INFO, "sshBase: %s", sshBase)

	clusterSize, err := dep.ClusterSize()
	if err != nil {
		return err
	}
	context.Logf(INFO, "clusterSize: %d", clusterSize)

	remoteDir := "/tmp"
	output := runSCPCommand(context, baseD, sd, sdReleaseFile, remoteDir, true)
	context.Logf(DEBUG, "SCP output: %s", output)

	// Run SSH command to call python:
	remotes := []string{remoteDir, path.Base(sdReleaseFile)}
	remoteFile := strings.Join(remotes, "/")
	sshCmd := append(sshBase, []string{
		"/usr/local/bin/stardog-update",
		baseD.Name,
		fmt.Sprintf("%d", clusterSize),
		remoteFile,
	}...)
	cmd := exec.Cmd{
		Path: sshCmd[0],
		Args: sshCmd,
	}
	context.Logf(DEBUG, "Running the update Stardog command: %s", strings.Join(sshCmd[:len(sshCmd)-1], " "))
	o, err := cmd.Output()
	if err != nil {
		context.Logf(ERROR, "Failed to update Stardog: %s", string(o))
		context.Logf(ERROR, "Error updating Stardog: %s", err)
		context.ConsoleLog(0, "Failed to update Stardog, verify that ssh agent is working and that the ssh key has been added to the agent.  Please run:\n")
		context.ConsoleLog(0, "\tssh-add %s\n", baseD.PrivateKey)
		return err
	}
	return nil
}

// GatherLogs sshes into the bastion node and collects logs from the stardog nodes
func GatherLogs(context AppContext, baseD *BaseDeployment, dep Deployment, outfile string) error {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return fmt.Errorf("ssh-agent needs to be setup for log gathering to work")
	}
	context.ConsoleLog(2, "Gathering logs...\n")
	sd, err := dep.FullStatus()
	if err != nil {
		return err
	}
	sshBase, err := getSSHCommand(context, baseD, sd)
	if err != nil {
		return err
	}
	clusterSize, err := dep.ClusterSize()
	if err != nil {
		return err
	}
	dstLogFile := fmt.Sprintf("/tmp/stardog%d.tar.gz", rand.Int())
	sshCmd := append(sshBase, []string{
		"/usr/local/bin/stardog-gather-logs",
		baseD.Name,
		fmt.Sprintf("%d", clusterSize),
		dstLogFile,
	}...)
	cmd := exec.Cmd{
		Path: sshCmd[0],
		Args: sshCmd,
	}
	context.Logf(DEBUG, "Running the log gathering command: %s", strings.Join(sshCmd[:len(sshCmd)-1], " "))
	o, err := cmd.Output()
	if err != nil {
		context.Logf(ERROR, "Failed to get the logs: %s", string(o))
		context.Logf(ERROR, "Error getting the logs: %s", err)
		context.ConsoleLog(0, "Failed to gather the logs, verify that ssh agent is working and that the ssh key has been added to the agent.  Please run:\n")
		context.ConsoleLog(0, "\tssh-add %s\n", baseD.PrivateKey)
		return err
	}
	context.Logf(DEBUG, "Log gathering output: %s", string(o))
	context.Logf(INFO, "Successfully gathered the logs on the bastion node at %s", dstLogFile)
	outfile = strings.TrimSpace(outfile)
	if outfile == "" {
		outfile = "stardoglogs.tar.gz"
	}
	return runSCPCommand(context, baseD, sd, outfile, dstLogFile, false)
}

// FullStatus inspects the state of a deployment and prints it out to the console.
func FullStatus(context AppContext, baseD *BaseDeployment, dep Deployment, internal bool, outfile string) error {
	context.ConsoleLog(2, "Checking status...\n")
	sd, err := dep.FullStatus()
	if err != nil {
		return err
	}

	sd.Healthy = IsHealthy(context, baseD, dep, internal)
	if sd.Healthy {
		context.ConsoleLog(1, "%s\n", context.SuccessString("The instance is healthy"))
	} else {
		context.ConsoleLog(1, "%s\n", context.FailString("The instance is not healthy"))
	}
	if os.Getenv("STARDOG_GRAVITON_UNIT_TEST") != "" {
		return nil
	}

	context.ConsoleLog(1, "Stardog is available here: %s\n", context.HighlightString(sd.StardogURL))
	context.ConsoleLog(1, "Stardog is internally available here: %s\n", context.HighlightString(sd.StardogInternalURL))
	context.ConsoleLog(1, "ssh is available here: %s\n", sd.SSHHost)

	pw := os.Getenv("STARDOG_ADMIN_PASSWORD")
	if pw == "" {
		pw = "admin"
	}

	client := stardogClientImpl{
		sdURL:    sd.StardogURL,
		logger:   context,
		username: "admin",
		password: pw,
	}
	nodes, err := client.GetClusterInfo()
	if err != nil {
		return err
	}
	context.ConsoleLog(1, "Using %d stardog nodes\n", len(*nodes))
	for _, n := range *nodes {
		context.ConsoleLog(1, "\t%s\n", n)
	}
	sd.StardogNodes = *nodes

	if outfile != "" {
		err = WriteJSON(sd, outfile)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
