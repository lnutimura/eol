package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FetchJSON(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Could not connect to %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request to %s failed with status: %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response from %s: %v", url, err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("Unable to parse response: %v", err)
	}

	return nil
}
