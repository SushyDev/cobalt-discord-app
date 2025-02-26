package cobalt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"io"
	"path/filepath"
	"strings"
)

// Client is a cobalt API client
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient creates a new cobalt API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: time.Second * 60, // Set a reasonable timeout
		},
	}
}

// ProcessVideo submits a video URL to the cobalt API for processing
func (c *Client) ProcessVideo(options RequestOptions) (*CobaltResponse, error) {
	// Create request payload
	payload := map[string]interface{}{
		"url":          options.URL,
		"videoQuality": options.VideoQuality,
		"downloadMode": options.DownloadMode,
	}
	
	// Add optional parameters if set
	if options.AudioFormat != "" {
		payload["audioFormat"] = options.AudioFormat
	}
	
	if options.AudioBitrate != "" {
		payload["audioBitrate"] = options.AudioBitrate
	}
	
	if options.FilenameStyle != "" {
		payload["filenameStyle"] = options.FilenameStyle
	}
	
	if options.YoutubeVideoCodec != "" {
		payload["youtubeVideoCodec"] = options.YoutubeVideoCodec
	}
	
	if options.YoutubeDubLang != "" {
		payload["youtubeDubLang"] = options.YoutubeDubLang
	}
	
	// Add boolean parameters (only if true)
	if options.AlwaysProxy {
		payload["alwaysProxy"] = true
	}
	
	if options.DisableMetadata {
		payload["disableMetadata"] = true
	}
	
	if options.TiktokFullAudio {
		payload["tiktokFullAudio"] = true
	}
	
	if options.TiktokH265 {
		payload["tiktokH265"] = true
	}
	
	if options.TwitterGif {
		payload["twitterGif"] = true
	}
	
	if options.YoutubeHLS {
		payload["youtubeHLS"] = true
	}
	
	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request payload: %w", err)
	}
	
	// Create request
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Add authorization header if API key is provided
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	}
	
	// Execute the request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check for non-success status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-success status code: %d, body: %s", resp.StatusCode, string(body))
	}
	
	// Unmarshal response
	var response CobaltResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	
	return &response, nil
}

// DownloadFile downloads a file from the given URL and returns it as a byte array
func (c *Client) DownloadFile(url, filename string) ([]byte, string, error) {
	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creating download request: %w", err)
	}
	
	// Execute the request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error executing download request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check for non-success status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("received non-success status code while downloading: %d", resp.StatusCode)
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error reading download response body: %w", err)
	}
	
	// If Content-Disposition header is present, extract filename from it
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" && filename == "" {
		// Parse filename from header
		// This is a simplified version, a proper implementation would need more robust parsing
		if start := strings.Index(contentDisposition, "filename="); start != -1 {
			start += 9
			end := strings.IndexRune(contentDisposition[start:], ';')
			if end == -1 {
				end = len(contentDisposition)
			} else {
				end += start
			}
			filename = contentDisposition[start:end]
			filename = strings.Trim(filename, `"'`)
		}
	}
	
	// If filename is still empty, extract from URL
	if filename == "" {
		filename = filepath.Base(url)
	}
	
	return body, filename, nil
}
