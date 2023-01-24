package internals

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/CopilotIQ/opa-tests/gentests"
	"io"
	"net/http"
	"strings"
)

const (
	contentType = "application/json"
	v1Data      = "v1/data"
	UrlSep      = "/"
)

func fullUrl(url string, endpoint string) string {
	return strings.Join([]string{url, v1Data, endpoint}, UrlSep)
}

func SendData(serverURL string, dataChan <-chan TestUnit) error {
	var failed = make([]string, 0)
	var total int
	for testUnit := range dataChan {
		total++
		jsonData, err := json.Marshal(testUnit.Body)
		if err != nil {
			return err
		}

		resp, err := http.Post(fullUrl(serverURL, testUnit.Endpoint), contentType,
			bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			failed = append(failed, testUnit.Name)
			resp.Body.Close()
			continue
		}
		b, err := GetResult(resp.Body)
		resp.Body.Close()
		if err != nil || b != testUnit.Expectation {
			failed = append(failed, testUnit.Name)
		}
		fmt.Print(".")
	}
	fmt.Printf(`
========================
Tests: %3d | Failed: %3d
------------------------
`, total, len(failed))
	for _, failTest := range failed {
		fmt.Println(failTest)
	}
	fmt.Println("========================")
	return nil
}

func GetResult(r io.Reader) (bool, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return false, err
	}
	result, ok := responseData["result"]
	if !ok {
		return false, fmt.Errorf("nothing to show")
	}
	b, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("%v is not a valid bool", result)
	}
	return b, nil
}
