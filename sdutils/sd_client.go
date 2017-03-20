package sdutils

import (
	"io"
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

type stardogClientImpl struct {
	sdURL    string
	password string
	username string
	logger   SdVaLogger
}


func (s *stardogClientImpl) doRequest(method, urlStr string, body io.Reader, contentType string, expectedCode int) ([]byte, error) {
	return s.doRequestWithAccept(method, urlStr, body, contentType, contentType, expectedCode)
}

func (s *stardogClientImpl) doRequestWithAccept(method, urlStr string, body io.Reader, contentType string, accept string, expectedCode int) ([]byte, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	client := &http.Client{}
	req.Header.Set("Content-Type", contentType)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed do the post %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedCode {
		return nil, fmt.Errorf("Expected %d but got %d when %s to %s", expectedCode, resp.StatusCode, method, urlStr)
	}
	content, err := ioutil.ReadAll(resp.Body)
	s.logger.Logf(DEBUG, "Completed %s to %s", method, urlStr)
	return content, nil
}

func (s *stardogClientImpl) GetClusterInfo() (*[]string, error) {
	s.logger.Logf(DEBUG, "GetClusterInfo\n")

	dbURL := fmt.Sprintf("%s/admin/cluster", s.sdURL)
	bodyBuf := &bytes.Buffer{}
	content, err := s.doRequest("GET", dbURL, bodyBuf, "application/json", 200)
	if err != nil {
		return nil, err
	}
	var nodesMap map[string]interface{}
	err = json.Unmarshal(content, &nodesMap)
	if err != nil {
		return nil, err
	}
	nodeList := nodesMap["nodes"]
	if nodeList == nil {
		return nil, fmt.Errorf("There is no available cluster information")
	}

	var ifaceList []interface{}
	switch v := nodeList.(type) {
	case []interface{}:
		s.logger.Logf(DEBUG, "Interface list %s", v)
		ifaceList = v
	default:
	// no match; here v has the same type as i
		return nil, fmt.Errorf("The returned cluster information was not expected %s", v)
	}

	outSList := make([]string, len(ifaceList))
	for i, nodeI := range ifaceList {
		outSList[i] = nodeI.(string)
	}
	return &outSList, nil
}