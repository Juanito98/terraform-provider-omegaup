package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL    string
	ApiToken   string
	HttpClient *http.Client
}

func NewClient(apiToken string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://omegaup.com"
	}
	client := &Client{
		BaseURL:    baseURL,
		ApiToken:   apiToken,
		HttpClient: http.DefaultClient,
	}

	return client
}

// Convert struct to map[string]string
func structToJson(obj interface{}) (map[string]string, error) {
	jsonData, err := json.Marshal(obj)
	var result map[string]string
	if err == nil {
		err = json.Unmarshal(jsonData, &result)
	}
	return result, err
}

func (c *Client) query(endpoint string, payload interface{}) ([]byte, error) {
	fullURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}
	payloadJson, err := structToJson(payload)
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range payloadJson {
		_ = writer.WriteField(key, val)
	}
	writer.Close()

	req, err := http.NewRequest("POST", fullURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.ApiToken))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("endpoint: %v, status: %v, error: %v", endpoint, resp.StatusCode, string(result))
	}

	return result, nil
}
