package workers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	. "github.com/CopilotIQ/opa-tests/common"
)

const contentType = "application/json"

func SendData(serverURL string, dataChan <-chan Request) error {
	for data := range dataChan {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		resp, err := http.Post(serverURL, contentType, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New("Server returned non-200 status code: " + resp.Status)
		}
	}
	return nil
}
