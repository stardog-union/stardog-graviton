package integration

import (
	//"os"
	"github.com/stardog-union/stardog-graviton"
	"math/rand"
	"fmt"
	"mime/multipart"
	"bytes"
	"net/http"
	"time"
	"io/ioutil"
	"os"
	"strings"
	"errors"
)

func getStardogUrl() (string, error) {
	stardogDescriptionPath := os.Getenv("STARDOG_DESCRIPTION_PATH")
	if stardogDescriptionPath == "" {
		return "", nil
	}
	var sdd sdutils.StardogDescription
	err := sdutils.LoadJSON(&sdd, stardogDescriptionPath)
	if err != nil {
		return "", err
	}
	return sdd.StardogURL, nil
}

func getRandomDbName(base string, n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("%s%s", base, string(b))
}

func makeDb(sdURL string, dbName string) error {
	data := fmt.Sprintf("{\"dbname\": \"%s\", \"options\" : {}, \"files\": []}", dbName)
	fmt.Printf("%s\n", data)

	dbUrl := fmt.Sprintf("%s/admin/databases", sdURL)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err := bodyWriter.WriteField("root", data)
	if err != nil {
		return fmt.Errorf("didnt make write field %s", err)
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, err := http.NewRequest("POST", dbUrl, bodyBuf)
	if err != nil {
		return fmt.Errorf("Failed to create the req %s url %s", dbUrl, err)
	}
	req.SetBasicAuth("admin", "admin")

	client := &http.Client{}
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed do the post %s", err)
	}
	if resp.StatusCode != 201 {
		fmt.Printf("ERROR %d %s\n", resp.ContentLength, resp.Status)
		return errors.New("Failed to create the database")
	}
	return nil
}

func startTransaction(sdURL string, dbName string) (string, error) {
	dbUrl := fmt.Sprintf("%s/%s/transaction/begin", sdURL, dbName)

	bodyBuf := &bytes.Buffer{}
	req, err := http.NewRequest("POST", dbUrl, bodyBuf)
	if err != nil {
		return "", fmt.Errorf("Failed to create the req %s url %s", dbUrl, err)
	}
	req.SetBasicAuth("admin", "admin")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed do the post %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to do the begin %s", resp.Status)
	}
	fmt.Printf("Begin Resp %s\n", resp.Status)
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read the id %s", err)
	}

	return string(content), nil
}

func postRows(sdURL string, dbName string, txId string, dataPath string) error {
	dbUrl := fmt.Sprintf("%s/%s/%s/add", sdURL, dbName, txId)

	f, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("Failed to open the file %s", dataPath)
	}

	req, err := http.NewRequest("POST", dbUrl, f)
	if err != nil {
		return fmt.Errorf("Failed to create the req %s url %s", dbUrl, err)
	}
	req.SetBasicAuth("admin", "admin")
	req.Header.Add("Content-Type", "application/rdf+xml")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed do the post %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to do the begin %s", resp.Status)
	}

	return nil
}

func commitTransaction(sdURL string, dbName string, txId string) error {
	dbUrl := fmt.Sprintf("%s/%s/transaction/commit/%s", sdURL, dbName, txId)

	bodyBuf := &bytes.Buffer{}
	req, err := http.NewRequest("POST", dbUrl, bodyBuf)
	if err != nil {
		return fmt.Errorf("Failed to create the req %s url %s", dbUrl, err)
	}
	req.SetBasicAuth("admin", "admin")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed do the post %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to do the begin %s", resp.Status)
	}

	return nil
}

func queryAll(sdURL string, dbName string) (string, error)  {
	data := "select * where { ?s ?p ?o } ORDER BY ?s"
	dbUrl := fmt.Sprintf("%s/%s/query", sdURL, dbName)

	fmt.Printf("%s\n", data)

	bodyBuf := strings.NewReader(data)
	contentType := "application/ld+json"

	req, err := http.NewRequest("POST", dbUrl, bodyBuf)
	if err != nil {
		return "", fmt.Errorf("Failed to create the req %s url %s", dbUrl, err)
	}
	req.SetBasicAuth("admin", "admin")

	client := &http.Client{}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed do the post %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("ERROR %d %s\n", resp.ContentLength, resp.Status)
		return "", fmt.Errorf("Query failed with mode %d", resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read the id %s", err)
	}
	// Pull of the head part that has node information
	return string(content[50:]), nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
