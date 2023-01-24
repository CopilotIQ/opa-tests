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

func SendData(serverURL string, dataChan <-chan TestUnit, report *TestReport) error {
	for testUnit := range dataChan {
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
			report.ReportFailure(testUnit.Name)
			resp.Body.Close()
			continue
		}
		b, err := GetResult(resp.Body)
		resp.Body.Close()
		if err != nil || b != testUnit.Expectation {
			report.ReportFailure(testUnit.Name)
		} else {
			report.IncSuccess()
		}
		fmt.Print(".")
	}
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
