package internals

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/CopilotIQ/opa-tests/testing"
	"io"
	"math"
	"net/http"
	"runtime"
	"strings"
	"sync"
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

func EstimateWorkers() uint {
	cores := runtime.NumCPU()
	if cores > 3 {
		Log.Debug("running with 70%% CPU load on %d cores", cores)
		return uint(math.Ceil(0.7 * float64(cores)))
	}
	return 1
}

func RunTests(tests []TestUnit, workers uint, addr string) *TestReport {
	dataChan := make(chan TestUnit)
	var wg sync.WaitGroup
	if workers == 0 {
		workers = EstimateWorkers()
	}
	if workers > 1 {
		Log.Info("running %d parallel test runners", workers)
	} else {
		Log.Warn("running single-core, execution will be slower")
	}
	url := fmt.Sprintf("http://%s", addr)
	var report TestReport
	for i := uint(0); i < workers; i++ {
		wg.Add(1)
		go func(num uint) {
			Log.Debug("starting worker #%d", num)
			err := SendData(url, dataChan, &report)
			if err != nil {
				Log.Error("error sending requests to OPA server: %v", err)
			}
			wg.Done()
			Log.Debug("worker #%d done", num)
		}(i)
	}
	for _, req := range tests {
		dataChan <- req
	}
	// Once you're done sending data, close the channel
	close(dataChan)
	wg.Wait()
	fmt.Println()
	return &report
}
