package internal

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Heartbeat(key string, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("AUTH", "Bearer"+key)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("response failed with status code: %d and body: %s", resp.StatusCode, body)
	}
	return nil
}
