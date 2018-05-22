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
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path/filepath"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

type validatorFunc func(key string) (interface{}, error)

// AskUser prompts a console user to enter input.  prompt is the string
// that will be displayed to them and defaultValue will be the result if
// the user just hits enter.
func AskUser(prompt string, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s (%s)", prompt, defaultValue)
	}
	fmt.Print(color.WhiteString("%s: ", prompt))
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	resultValue := strings.TrimSpace(response)
	if resultValue == "" {
		resultValue = defaultValue
	}
	return resultValue, nil
}

// AskUserYesOrNo is just a convenience wrapper around AskUser that looks for
// a yes or no answer.  A case insensitive yes will return true and all other
// values will return false.
func AskUserYesOrNo(prompt string) bool {
	res, err := AskUser(prompt, "yes/no")
	if err != nil {
		return false
	}
	return strings.ToLower(res) == "yes"
}

func askUserInteractive(prompt string, defaultValue string, skipIfDefault bool, vf validatorFunc) (interface{}, error) {
	if skipIfDefault && defaultValue != "" {
		return vf(defaultValue)
	}
	v, err := AskUser(prompt, defaultValue)
	if err != nil {
		return nil, err
	}
	return vf(v)
}

// AskUserInteractiveInt prompts the user to enter an integer.
func AskUserInteractiveInt(prompt string, defaultValue int, skipIfDefault bool, val *int) error {
	v, err := askUserInteractive(prompt, fmt.Sprintf("%d", defaultValue), skipIfDefault, stringToIntegerValidator)
	if err != nil {
		return err
	}
	*val = v.(int)
	return nil
}

// AskUserInteractiveString prompts the user to enter a string.
func AskUserInteractiveString(prompt string, defaultValue string, skipIfDefault bool, val *string) error {
	c := ""
	for c == "" {
		v, err := askUserInteractive(prompt, defaultValue, skipIfDefault, stringToStringValidator)
		if err != nil {
			return err
		}
		c = v.(string)
		if c == "" {
			fmt.Print("A value must be provided.\n")
		}
	}
	*val = c
	return nil
}

func stringToIntegerValidator(key string) (interface{}, error) {
	return strconv.Atoi(key)
}

func stringToStringValidator(key string) (interface{}, error) {
	return key, nil
}

func GetMaxIopsRatio() (int) {
	return 50
}

// Spinner is an object used to show progress on the console.
type Spinner struct {
	nextMap  map[string]string
	lastSpin string
	message  string
	level    int
	context  AppContext
}

// NewSpinner creates a new spinner object.
func NewSpinner(context AppContext, level int, message string) *Spinner {
	s := Spinner{
		lastSpin: "|",
		message:  message,
		level:    level,
		context:  context,
	}
	s.nextMap = map[string]string{
		"|":  "/",
		"/":  "-",
		"-":  "\\",
		"\\": "|",
	}
	return &s
}

// EchoNext prints out the progress character.
func (s *Spinner) EchoNext() {
	if color.NoColor {
		s.context.ConsoleLog(s.level, ".")
	} else {
		s.lastSpin = s.nextMap[s.lastSpin]
		s.context.ConsoleLog(s.level, "\r%s %s...", s.context.HighlightString(s.lastSpin), s.message)
	}
}

// Close ends the spinner session.
func (s *Spinner) Close() {
	s.context.ConsoleLog(s.level, "\n")
}

// ScanResult is an object returned from a LineScanner.  This allows us to use
// the uility function RunScanner and return different values from the output
// based on the specific command.
type ScanResult struct {
	Key   string
	Value string
}

// LineScanner is a function that will search a line for given values and return
// results in a ScanResult if it finds something.  It may return nil
type LineScanner func(cliContext AppContext, line string) *ScanResult

// RunCommand will fork and execute a command in the shell.  The lineScanner object
// will be used to collect output and return it to the caller.
func RunCommand(cliContext AppContext, cmd exec.Cmd, lineScanner LineScanner, spinner *Spinner) (*[]ScanResult, error) {
	cliContext.Logf(DEBUG, "Start the program %s\n", cmd.Args[0])

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdErrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdErrScanner.Scan() {
			text := stdErrScanner.Text()
			cliContext.Logf(WARN, "STDERR %s", text)
		}
	}()

	if err := cmd.Start(); err != nil {
		cliContext.Logf(WARN, "Failed to start the program %s\n", cmd.Args[0])
		return nil, err
	}
	cliContext.Logf(DEBUG, "Started the program %s\n", cmd.Args[0])

	var scanResults []ScanResult
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if spinner != nil {
			spinner.EchoNext()
		}
		cliContext.Logf(DEBUG, "STDOUT %s", line)
		if lineScanner != nil {
			rc := lineScanner(cliContext, line)
			if rc != nil {
				scanResults = append(scanResults, *rc)
			}
		}
	}
	if spinner != nil {
		spinner.Close()
	}

	cliContext.Logf(DEBUG, "Waiting for the program %s to complete\n", cmd.Args[0])
	err = cmd.Wait()
	if err != nil {
		cliContext.Logf(ERROR, "The program %s failed to complete: %s\n", cmd.Args[0], err)
		return nil, err
	}
	cliContext.Logf(DEBUG, "The program %s completed\n", cmd.Args[0])

	if !cmd.ProcessState.Success() {
		return nil, fmt.Errorf("Command failed %s", cmd.Args[0])
	}
	return &scanResults, nil
}

func WriteFile(path string, contents string) error {
	data := []byte(contents)
	err := ioutil.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

// WriteJSON will take an interface object and serialize it into JSON and store it
// in a file at the given path.
func WriteJSON(obj interface{}, path string) error {
	data, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

// LoadJSON is a convenience function to load a JSON file into an interface object
func LoadJSON(obj interface{}, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("The file %s does not exist", path)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	return nil
}

// PathExists is a convenience function to determine if a path path exists.
func PathExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

// BbCode converts the bb ascii art information into console colorsMap
func BbCode(data string) {
	colorsMap := make(map[string]*color.Color)
	colorsMap["00aa50"] = color.New(color.FgGreen)
	colorsMap["00aa00"] = color.New(color.FgGreen)
	colorsMap["55ffff"] = color.New(color.FgCyan)
	colorsMap["aa5500"] = color.New(color.FgYellow)
	colorsMap["55ff55"] = color.New(color.FgHiGreen)
	colorsMap["aa0000"] = color.New(color.FgBlack)
	colorsMap["0000aa"] = color.New(color.FgBlue)
	colorsMap["ffffff"] = color.New(color.FgWhite)
	colorsMap["ff55ff"] = color.New(color.FgHiRed)
	colorsMap["555555"] = color.New(color.FgWhite)
	colorsMap["aaaaaa"] = color.New(color.FgHiWhite)
	colorsMap["5555ff"] = color.New(color.FgBlue)
	colorsMap["ffff55"] = color.New(color.FgYellow)
	colorsMap["ff5555"] = color.New(color.FgRed)

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		ptr := line
		ndx := strings.Index(ptr, "[color=#")
		for ndx > -1 {
			ptr = ptr[ndx:]
			colorString := ptr[8:14]
			endNdx := strings.Index(ptr, "[/color]")
			line := ptr[15:endNdx]
			ptr = ptr[endNdx+7:]
			c := colorsMap[colorString]
			c.Printf("%s", line)
			ndx = strings.Index(ptr, "[color=#")
		}
		fmt.Println()
	}
}

// GetLocalOnlyHTTPMask uses a network service to guess the external IP of the local host.
func GetLocalOnlyHTTPMask() string {
	url := "http://ip.42.pl/raw"
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ""
	}
	cidr := fmt.Sprintf("%s/32", string(b))
	return cidr
}

func GenerateKey(dir string, keyname string) (string, []byte, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", nil, err
	}

	privateKeyFilename := filepath.Join(dir, keyname)
	pubKeyFilename := privateKeyFilename + ".pub"

	if PathExists(privateKeyFilename) {
		return "", nil, fmt.Errorf("The private key %s already exists", pubKeyFilename)
	}
	if PathExists(pubKeyFilename) {
		return "", nil, fmt.Errorf("The private key %s already exists", pubKeyFilename)
	}

	privateKeyFile, err := os.OpenFile(privateKeyFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer privateKeyFile.Close()
	if err != nil {
		return "", nil, err
	}

	pemKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)}
	if err := pem.Encode(privateKeyFile, pemKey); err != nil {
		return "", nil, err
	}
	pub, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return "", nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(pub)
	err = ioutil.WriteFile(pubKeyFilename, pubKeyBytes, 0655)
	if err != nil {
		return "", nil, err
	}

	return privateKeyFilename, pubKeyBytes, nil
}

// ValueStringToInt returns a integer from a string with a unit of g, m, or k.
func ValueStringToInt(i string) (int, error) {
	multTab := make(map[string]int)
	multTab["g"] = 1024 * 1024 * 1024
	multTab["m"] = 1024 * 1024
	multTab["k"] = 1024 * 1024

	unit := i[len(i)-1 : len(i)]
	mult, ok := multTab[strings.ToLower(unit)]
	if ok {
		i = i[0 : len(i)-1]
	} else {
		mult = 1
	}
	base, err := strconv.Atoi(i)
	if err != nil {
		return 0, fmt.Errorf("%s is not valid", i)
	}
	return base * mult, nil
}

func checkVersion(context AppContext, program string, version string) bool {
	versionOpt := "--version"
	cmd := exec.Command(program, versionOpt)
	out, err := cmd.Output()
	if err != nil {
		context.Logf(INFO, "The program %s unsuccessfully ran with the %s option", program, versionOpt)
		return false
	}
	outA := strings.Split(string(out), "\n")
	if len(outA) < 1 {
		return false
	}
	return strings.TrimSpace(outA[0]) == version
}

func unzipFile(context AppContext, zipPath string) error {
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		destPath := filepath.Join(context.GetConfigDir(), f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(destPath), f.Mode())
			f, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	zipR, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	for _, f := range zipR.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadHTTPFile(context AppContext, destPath string, url string) error {
	// We have to write the data to a file because the zip module requires seeking
	zipF, err := ioutil.TempFile("", "graviton_zip_dl")
	if err != nil {
		return err
	}
	defer os.Remove(zipF.Name())

	resp, err := http.Get(url)
	if err != nil {
		zipF.Close()
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(zipF, resp.Body)
	if err != nil {
		zipF.Close()
		return err
	}
	zipF.Close()

	err = unzipFile(context, zipF.Name())
	if err != nil {
		return err
	}
	return nil
}

// CopyFile will copy the context of 1 path to another
func CopyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Sync()
	if err != nil {
		return err
	}
	return nil
}

// FindProgramVersion is used to check for golang style single executable programs of a
// specific version.  We are currently using it for just packer and terraform and there
// are no guarantees that it will work for other programs with different path expectations
func FindProgramVersion(context AppContext, program string, version string, url string) error {
	// First check if the program is already in the graviton environment
	gravPgm := filepath.Join(context.GetConfigDir(), program)
	if PathExists(gravPgm) {
		context.Logf(INFO, "%s exists in graviton directory.  Checking for version %s\n", program, version)
		if checkVersion(context, gravPgm, version) {
			context.Logf(INFO, "%s has version %s", program, version)
			return nil
		}
	}
	// Now search the users path for the program
	systemPgm, err := exec.LookPath(program)
	if err == nil {
		context.Logf(INFO, "%s exists in the system path.  Checking for version %s\n", program, version)
		if checkVersion(context, systemPgm, version) {
			err = CopyFile(systemPgm, gravPgm)
			if err != nil {
				return err
			}
			err = os.Chmod(gravPgm, 0777)
			if err != nil {
				return err
			}
			context.Logf(INFO, "%s has version %s\n", program, version)
			return nil
		}
	}
	context.Logf(INFO, "Downloading %s\n", url)
	err = downloadHTTPFile(context, gravPgm, url)
	if err == nil {
		if checkVersion(context, gravPgm, version) {
			context.Logf(INFO, "%s has version %s\n", program, version)
			return nil
		}
	} else {
		context.Logf(INFO, "Failed to download %s: %s\n", url, err.Error())
	}
	return fmt.Errorf("We were unable to find a valid %s, please install version %s on your system", program, version)
}
